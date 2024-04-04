package bhlnames

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/gnames/bayes"
	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/builder"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/nlp"
	"github.com/gnames/bhlnames/internal/ent/reffnd"
	"github.com/gnames/bhlnames/internal/ent/score"
	"github.com/gnames/bhlnames/internal/ent/ttlmch"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnparser"
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

	res.gnpPool = gnparserPool(cfg.JobsNum)
	return &res
}

// Initialize downloads BHL's metadata and imports it into the storage.
func (bn bhlnames) Initialize(bld builder.Builder) error {
	var err error
	if bn.cfg.WithRebuild {
		bld.ResetData()
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

	return bn.Close()
}

func (bn bhlnames) NameRefs(inp input.Input) (*bhl.RefsByName, error) {
	res, err := bn.rf.ReferencesByName(inp, bn.cfg)
	if err != nil {
		return nil, err
	}
	res.ReferenceNumber = len(res.References)

	if inp.WithNomenEvent || inp.Reference.RefString != "" {
		bn.sortByScore(res, inp.WithNomenEvent)
	}

	if inp.RefsLimit > 0 && len(res.References) > inp.RefsLimit {
		res.References = res.References[:inp.RefsLimit]
	}

	if inp.Reference.RefString != "" {
		for i := range res.References {
			res.References[i].RefMatchQuality = matchQuality(res.References[i].Odds)
		}
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

func (bn bhlnames) NameRefsStream(
	chIn <-chan input.Input,
	chOut chan<- *bhl.RefsByName,
) {
}

func (bn bhlnames) Config() config.Config {
	return bn.cfg
}

func (bn bhlnames) ParserPool() chan gnparser.GNparser {
	return bn.gnpPool
}

func (bn *bhlnames) Close() error {
	if bn.rf != nil {
		bn.rf.Close()
	}
	if bn.tm != nil {
		err := bn.tm.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func gnparserPool(poolSize int) chan gnparser.GNparser {
	gnpPool := make(chan gnparser.GNparser, poolSize)
	for range poolSize {
		cfgGNP := gnparser.NewConfig()
		gnpPool <- gnparser.New(cfgGNP)
	}
	return gnpPool
}

func (bn bhlnames) sortByScore(nr *bhl.RefsByName, isNomen bool) error {
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

	if len(nr.References) > 0 {
		noScoreIndex := len(nr.References)
		for i := range nr.References {
			if nr.References[i].Score.Total == 0 {
				noScoreIndex = i
				break
			}
		}

		nr.References = nr.References[:noScoreIndex]
	}

	err = score.BoostBestResult(nr, bn.bs)
	if err != nil {
		return err
	}

	return nil
}
