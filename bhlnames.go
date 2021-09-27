package bhlnames

import (
	"log"
	"sort"
	"sync"

	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/builder"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/reffinder"
	"github.com/gnames/bhlnames/ent/score"
	"github.com/gnames/gnfmt"
)

type bhlnames struct {
	cfg config.Config
}

func New(cfg config.Config) BHLnames {
	bn := &bhlnames{cfg: cfg}
	return bn
}

func (bn *bhlnames) Initialize(b builder.Builder) error {
	if bn.cfg.WithRebuild {
		b.ResetData()
	}
	return b.ImportData()
}

func (bn *bhlnames) NameRefs(
	rf reffinder.RefFinder,
	data input.Input,
) (*namerefs.NameRefs, error) {
	return rf.ReferencesBHL(data)
}

func (bn *bhlnames) NameRefsStream(
	rf reffinder.RefFinder,
	chIn <-chan input.Input,
	chOut chan<- *namerefs.NameRefs,
) {
	var wg sync.WaitGroup
	wg.Add(bn.cfg.JobsNum)

	for i := 0; i < bn.cfg.JobsNum; i++ {
		go bn.nameRefsWorker(rf, chIn, chOut, &wg)
	}
	wg.Wait()
	close(chOut)
}

func (bn *bhlnames) nameRefsWorker(
	rf reffinder.RefFinder,
	chIn <-chan input.Input,
	chOut chan<- *namerefs.NameRefs,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for data := range chIn {
		nameRefs, err := rf.ReferencesBHL(data)
		if err != nil {
			log.Println(err)
		}
		chOut <- nameRefs
	}
}

func (bn *bhlnames) NomenRefs(
	rf reffinder.RefFinder,
	inp input.Input,
) (*namerefs.NameRefs, error) {
	nr, err := rf.ReferencesBHL(inp)
	if err != nil {
		return nil, err
	}

	sortByScore(nr)
	return nr, nil
}

func (bn *bhlnames) NomenRefsStream(
	rf reffinder.RefFinder,
	chIn <-chan input.Input,
	chOut chan<- *namerefs.NameRefs,
) {
	var wg sync.WaitGroup
	wg.Add(bn.cfg.JobsNum)

	for i := 0; i < bn.cfg.JobsNum; i++ {
		go bn.nomenRefsWorker(rf, chIn, chOut, &wg)
	}
	wg.Wait()
	close(chOut)
}

func (bn *bhlnames) nomenRefsWorker(
	rf reffinder.RefFinder,
	chIn <-chan input.Input,
	chOut chan<- *namerefs.NameRefs,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for data := range chIn {
		nr, err := rf.ReferencesBHL(data)
		if err != nil {
			log.Println(err)
		}
		sortByScore(nr)
		chOut <- nr
	}
}

func sortByScore(nr *namerefs.NameRefs) {
	// Year has precedence over others
	prec := map[score.ScoreType]int{score.Annot: 0, score.Year: 1}
	s := score.New(prec)
	s.Calculate(nr)
	sort.Slice(nr.References, func(i, j int) bool {
		refs := nr.References
		if refs[i].Score.Sort == refs[j].Score.Sort {
			return refs[i].YearAggr < refs[j].YearAggr
		}
		return refs[i].Score.Sort > refs[j].Score.Sort
	})
	// if len(nr.References) > 0 {
	// 	nr.References = nr.References[:1]
	// }
}

func (bn *bhlnames) Format() gnfmt.Format {
	return bn.cfg.Format
}
