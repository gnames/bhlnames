package bhlnames

import (
	"cmp"
	"fmt"
	"slices"
	"sync"

	"github.com/gnames/bayes"
	"github.com/gnames/bhlnames/internal/ent/builder"
	"github.com/gnames/bhlnames/internal/ent/colbuild"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/bhlnames/internal/ent/nlp"
	"github.com/gnames/bhlnames/internal/ent/refbhl"
	"github.com/gnames/bhlnames/internal/ent/reffinder"
	"github.com/gnames/bhlnames/internal/ent/score"
	"github.com/gnames/bhlnames/internal/ent/title_matcher"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnlib/ent/gnvers"
	"github.com/gnames/gnparser"
	"github.com/rs/zerolog/log"
)

type Option func(*bhlnames)

func OptColBuild(c colbuild.ColBuild) Option {
	return func(bn *bhlnames) {
		bn.ColBuild = c
	}
}

func OptBuilder(b builder.Builder) Option {
	return func(bn *bhlnames) {
		bn.Builder = b
	}
}

func OptRefFinder(rf reffinder.RefFinder) Option {
	return func(bn *bhlnames) {
		bn.RefFinder = rf
	}
}

func OptTitleMatcher(tm title_matcher.TitleMatcher) Option {
	return func(bn *bhlnames) {
		bn.TitleMatcher = tm
	}
}

func OptParser(gnp gnparser.GNparser) Option {
	return func(bn *bhlnames) {
		bn.GNparser = gnp
	}
}

func OptNLP(n nlp.NLP) Option {
	return func(bn *bhlnames) {
		bayesWithData := n.LoadPretrainedWeights()
		bn.Bayes = bayesWithData
	}
}

type bhlnames struct {
	cfg config.Config
	gnparser.GNparser
	builder.Builder
	colbuild.ColBuild
	reffinder.RefFinder
	title_matcher.TitleMatcher
	bayes.Bayes
}

// New creates an instance of BHLnames.
func New(cfg config.Config, opts ...Option) BHLnames {
	bn := bhlnames{cfg: cfg}
	for i := range opts {
		opts[i](&bn)
	}
	return &bn
}

// Close Terminates connections to databases and key-value stores.
func (bn bhlnames) Close() error {
	var err error
	if bn.RefFinder != nil {
		err = bn.RefFinder.Close()
	}
	if err == nil && bn.TitleMatcher != nil {
		err = bn.TitleMatcher.Close()
	}
	return err
}

// Parser returns an instance of the Global Names Parser. It is used
// to parse scientific names.
func (bn bhlnames) Parser() gnparser.GNparser {
	return bn.GNparser
}

// Initialize downloads BHL's metadata and imports it into the storage.
func (bn bhlnames) Initialize() error {
	var err error
	if bn.cfg.WithRebuild {
		bn.ResetData()
	}

	err = bn.ImportData()
	if err != nil {
		err = fmt.Errorf("ImportData: %w", err)
		return err
	}

	err = bn.CalculateTxStats()
	if err != nil {
		err = fmt.Errorf("CalculateTxStats: %w", err)
	}
	return err
}

// InitializeCol downloads the Catalogue of Life (CoL) data and imports
// the data about nomenclatural references for CoL names into the storage.
func (bn bhlnames) InitializeCol() error {
	var err error
	var hasFiles, hasData bool
	hasFiles, hasData, err = bn.DataStatus()
	if err != nil {
		err = fmt.Errorf("DataStatus: %w", err)
		return err
	}

	if bn.cfg.WithRebuild || !hasFiles {
		bn.ResetColData()
	}

	// do not reimport data if percentage is given
	if bn.cfg.WithCoLRecalc || !hasData {
		err = bn.ImportColData()
		if err != nil {
			err = fmt.Errorf("ImportColData: %w", err)
			return err
		}
	}

	err = bn.LinkColToBhl(bn.NomenRefsStream)
	if err != nil {
		err = fmt.Errorf("LinkColToBhl: %w", err)
	}
	return err
}

// RefByPageID returns a reference metadata for a given pageID.
func (bn bhlnames) RefByPageID(pageID int) (*refbhl.Reference, error) {
	return bn.RefFinder.RefByPageID(pageID)
}

// NameRefs takes a name and optionally reference, and find matching
// locations and references in BHL.
func (bn bhlnames) NameRefs(inp input.Input) (*namerefs.NameRefs, error) {
	return bn.ReferencesBHL(inp, bn.cfg)
}

// NameRefsStream takes a stream of names/references and returns back
// a stream of matched locations in BHL.
func (bn bhlnames) NameRefsStream(
	chIn <-chan input.Input,
	chOut chan<- *namerefs.NameRefs,
) {
	var wg sync.WaitGroup
	wg.Add(bn.cfg.JobsNum)

	for i := 0; i < bn.cfg.JobsNum; i++ {
		go bn.nameRefsWorker(chIn, chOut, &wg)
	}
	wg.Wait()
	close(chOut)
}

func (bn bhlnames) nameRefsWorker(
	chIn <-chan input.Input,
	chOut chan<- *namerefs.NameRefs,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for inp := range chIn {
		nameRefs, err := bn.ReferencesBHL(inp, bn.cfg)
		if err != nil {
			err = fmt.Errorf("bhlnames.nameRefsWorker: %w", err)
			log.Warn().Err(err)
		}
		chOut <- nameRefs
	}
}

func (bn bhlnames) NomenRefs(
	inp input.Input,
) (*namerefs.NameRefs, error) {
	nrs, err := bn.ReferencesBHL(inp, bn.cfg)
	if err != nil {
		return nil, err
	}
	err = bn.sortByScore(nrs)
	return nrs, err
}

func (bn bhlnames) NomenRefsStream(
	chIn <-chan input.Input,
	chOut chan<- *namerefs.NameRefs,
) {
	var wg sync.WaitGroup
	wg.Add(bn.cfg.JobsNum)

	for i := 0; i < bn.cfg.JobsNum; i++ {
		go bn.nomenRefsWorker(chIn, chOut, &wg)
	}
	wg.Wait()
	close(chOut)
}

func (bn bhlnames) nomenRefsWorker(
	chIn <-chan input.Input,
	chOut chan<- *namerefs.NameRefs,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for inp := range chIn {
		nr, err := bn.ReferencesBHL(inp, bn.cfg)
		if err != nil {
			err = fmt.Errorf("bhlnames.nomenRefsWorker: %#w", err)
			log.Warn().Err(err)
		}
		err = bn.sortByScore(nr)
		if err != nil {
			err = fmt.Errorf("bhlnames.nomenRefsWorker: %#w", err)
			log.Warn().Err(err)
		}
		chOut <- nr
	}
}

func (bn bhlnames) sortByScore(nr *namerefs.NameRefs) error {
	// Year has precedence over others
	prec := map[score.ScoreType]int{
		score.RefVolume: 0,
		score.RefTitle:  1,
		score.Annot:     2,
		score.Year:      3,
		score.RefPages:  4,
	}
	s := score.New(prec)
	err := s.Calculate(nr, bn.TitleMatcher, bn.Bayes)
	if err != nil {
		return err
	}
	slices.SortFunc(nr.References, func(a, b *refbhl.ReferenceNameBHL) int {
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

	score.BoostBestResult(nr, bn.Bayes)
	return nil
}

func (bn bhlnames) Config() config.Config {
	return bn.cfg
}

func (bn bhlnames) ChangeConfig(opts ...config.Option) BHLnames {
	for _, o := range opts {
		o(&bn.cfg)
	}
	return bn
}

func (bn bhlnames) GetVersion() gnvers.Version {
	return gnvers.Version{
		Version: Version,
		Build:   Build,
	}
}
