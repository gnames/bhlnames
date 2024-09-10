package bhlnames

import (
	"cmp"
	"fmt"
	"log/slog"
	"slices"

	"github.com/gnames/bayes"
	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/builder"
	"github.com/gnames/bhlnames/internal/ent/col"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/nlp"
	"github.com/gnames/bhlnames/internal/ent/reffnd"
	"github.com/gnames/bhlnames/internal/ent/score"
	"github.com/gnames/bhlnames/internal/ent/ttlmch"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnsys"
)

// Option provides an 'interface' for setting up BHLnames instance.
type Option func(*bhlnames)

// OptRefFinder sets the RefFinder for finding references in BHL.
func OptRefFinder(rf reffnd.RefFinder) Option {
	return func(bn *bhlnames) {
		bn.rf = rf
	}
}

// OptTitleMatcher sets the TitleMatcher for finding possible matches to a
// reference title from the input.
func OptTitleMatcher(tm ttlmch.TitleMatcher) Option {
	return func(bn *bhlnames) {
		bn.tm = tm
	}
}

// OptNLP creates NLP instance to generate Odds for references.
func OptNLP(n nlp.NLP) Option {
	return func(bn *bhlnames) {
		bayesWithData := n.LoadPretrainedWeights()
		bn.bs = bayesWithData
	}
}

// bhlnames implements BHLnames interface.
type bhlnames struct {
	// cfg is a configuration for BHLnames.
	cfg config.Config

	// rf is a RefFinder for BHLnames. RefFinder is used for finding
	// references in BHL according to input.
	rf reffnd.RefFinder

	// tm is a TitleMatcher for BHLnames. TitleMatcher is used for
	// finding possible matches to a reference title from the input.
	tm ttlmch.TitleMatcher

	// Bayes is a bayesian classifier trained on nomenclatural events papers
	// found in BHL.
	bs bayes.Bayes

	// gnPool is a pool of gnparser instances. Thy are used for the scientific
	// name parsing.
	gnpPool chan gnparser.GNparser
}

// New creates a new BHLnames instance.
func New(cfg config.Config, opts ...Option) BHLnames {
	res := bhlnames{cfg: cfg}
	for _, opt := range opts {
		opt(&res)
	}

	res.gnpPool = gnparser.NewPool(gnparser.NewConfig(), cfg.JobsNum)
	return &res
}

// Initialize downloads BHL's metadata and imports it into the storage.
func (bn bhlnames) Initialize(bld builder.Builder) error {
	var err error
	defer bn.Close()

	if bn.cfg.WithRebuild {
		bld.ResetData()
	} else {
		err = gnsys.MakeDir(bn.cfg.RootDir)
		if err != nil {
			slog.Error(
				"Cannot create root directory",
				"dir", bn.cfg.RootDir,
				"error", err,
			)
			return err
		}
	}

	err = bld.ImportData()
	if err != nil {
		err = fmt.Errorf("ImportData: %w", err)
		return err
	}

	err = bld.CalculateTxStats()
	if err != nil {
		err = fmt.Errorf("CalculateTxStats: %w", err)
		return err
	}

	return nil
}

func (bn bhlnames) NameRefs(inp input.Input) (*bhl.RefsByName, error) {
	res, err := bn.rf.ReferencesByName(inp, bn.cfg)
	if err != nil {
		return nil, err
	}
	// do not show ReferenceNumber for nomenclatural events, because we
	// try to find only one reference.
	if inp.WithNomenEvent {
		res.ReferenceNumber = 0
	} else {
		res.ReferenceNumber = len(res.References)
	}

	// limit the number of references if needed
	if inp.RefsLimit > 0 && len(res.References) > inp.RefsLimit {
		res.References = res.References[:inp.RefsLimit]
	}

	// if results are from CoL cache, return them here
	if res.Meta.NomenEventFromCache {
		return res, nil
	}

	if inp.WithNomenEvent || inp.Reference != nil {
		bn.scoreCalcSort(res, inp.WithNomenEvent)
	}

	if inp.Reference != nil {
		for i := range res.References {
			res.References[i].RefMatchQuality = matchQuality(res.References[i].Odds)
		}
	}

	return res, nil
}

// RefByPageID returns a reference metadata for a given pageID.
func (bn bhlnames) RefByPageID(pageID int) (*bhl.Reference, error) {
	return bn.rf.RefByPageID(pageID)
}

func (bn bhlnames) RefsByExtID(
	extID string,
	dataSourceID int,
	allRefs bool,
) (*bhl.RefsByName, error) {
	res, err := bn.rf.RefsByExtID(extID, dataSourceID)
	if err != nil {
		return nil, err
	}

	if !allRefs && res != nil {
		if len(res.References) > 0 {
			res.References = res.References[:1]
		}
	}
	return res, nil
}

func (bn bhlnames) ItemStats(itemID int) (*bhl.Item, error) {
	res, err := bn.rf.ItemStats(itemID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (bn bhlnames) ItemsByTaxon(taxon string) ([]*bhl.Item, error) {
	res, err := bn.rf.ItemsByTaxon(taxon)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func matchQuality(odds float64) int {
	if odds <= 0 {
		return 0
	}
	if odds < 0.01 {
		return 1
	}
	if odds < 0.1 {
		return 2
	}
	if odds < 1.0 {
		return 3
	}
	if odds < 10.0 {
		return 4
	}
	return 5
}

func (bn bhlnames) InitCoLNomenEvents(cn col.Nomen) error {
	hasFiles, hasData, err := cn.CheckCoLData()
	if err != nil {
		slog.Error("Unable to check Catalogue of Life data", "error", err)
		return err
	}

	if bn.cfg.WithRebuild || !hasFiles {
		cn.ResetCoLData()
	}

	// do not reimport data if percentage is given
	if bn.cfg.WithCoLDataTrim || !hasData {
		err = cn.ImportCoLData()
		if err != nil {
			slog.Error("Unable to import Catalogue of Life data", "error", err)
			return err
		}
	}

	err = cn.NomenEvents(bn.NameRefsStream)
	if err != nil {
		slog.Error("Unable to get nomenclatural events for CoL", "error", err)
		return err
	}
	return nil
}

func (bn bhlnames) Config() config.Config {
	return bn.cfg
}

func (bn bhlnames) ParserPool() chan gnparser.GNparser {
	return bn.gnpPool
}

func (bn *bhlnames) Close() {
	if bn.rf != nil {
		bn.rf.Close()
	}
	if bn.tm != nil {
		bn.tm.Close()
	}
}

func (bn bhlnames) scoreCalcSort(nr *bhl.RefsByName, isNomen bool) error {
	// Year has precedence over others
	prec := map[score.ScoreType]int{
		score.RefVolume: 0,
		score.RefTitle:  1,
		score.Annot:     2,
		score.Year:      3,
		score.RefPages:  4,
	}
	s := score.New(prec)
	err := s.Calculate(nr, bn.tm, bn.bs, isNomen)
	if err != nil {
		return err
	}
	slices.SortFunc(nr.References, func(a, b *bhl.ReferenceName) int {
		if a.Score.Odds == b.Score.Odds {
			if a.YearAggr == b.YearAggr {
				return cmp.Compare(a.PageID, b.PageID)
			}
			return cmp.Compare(a.YearAggr, b.YearAggr)
		}
		return cmp.Compare(b.Score.Odds, a.Score.Odds)
	})

	// if len(nr.References) > 0 {
	// 	noScoreIndex := len(nr.References)
	// 	for i := range nr.References {
	// 		if nr.References[i].Score.Total == 0 {
	// 			noScoreIndex = i
	// 			break
	// 		}
	// 	}
	//
	// 	nr.References = nr.References[:noScoreIndex]
	// }

	err = score.BoostBestResult(nr, bn.bs)
	if err != nil {
		return err
	}

	return nil
}
