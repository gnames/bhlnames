package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/refbhl"
	"github.com/gnames/bhlnames/ent/score"
	"github.com/gnames/bhlnames/io/bayesio"
	"github.com/gnames/bhlnames/io/titlemio"
	"github.com/gnames/gnfmt"
	"github.com/rs/zerolog/log"
)

var dataPath = filepath.Join("..", "..", "io", "bayesio", "data")
var goldFile = filepath.Join(dataPath, "gold.json")
var resFile = "res.json"

const namesTotal = 3000

var maxOdds float64

type stats struct {
	namesNum     int
	nameNomenNum int
	refs         int
	refsNomen    int
	distr
}

type distr struct {
	score100, score10, score1, score01, score001, score0 distData
}

type distData struct {
	isNomen, notNomen int
}

func main() {
	nlp := bayesio.New()
	nb := nlp.Load()
	cfg := config.New()
	tm := titlemio.New(cfg)
	defer tm.Close()

	gold, err := os.ReadFile(goldFile)
	if err != nil {
		log.Fatal().Err(err)
	}
	var nrs []*namerefs.NameRefs
	enc := gnfmt.GNjson{Pretty: true}
	err = enc.Decode(gold, &nrs)
	if err != nil {
		log.Fatal().Err(err)
	}
	sts := new(stats)
	sts.namesNum = len(nrs)
	_ = sts
	for i := range nrs {
		var hasNomenRef bool
		if nrs[i].References[0].IsNomenRef {
			hasNomenRef = true
			sts.nameNomenNum++
		}
		prec := map[score.ScoreType]int{
			score.RefVolume: 0,
			score.RefTitle:  1,
			score.Annot:     2,
			score.Year:      3,
			score.RefPages:  4,
		}
		sc := score.New(prec)
		sc.Calculate(nrs[i], tm, nb)
		if nrs[i].References[0].Score.Odds >= 100 {
			if hasNomenRef {
				sts.distr.score100.isNomen++
			} else {
				sts.distr.score100.notNomen++
			}
		} else if nrs[i].References[0].Score.Odds >= 10 {
			if hasNomenRef {
				sts.distr.score10.isNomen++
			} else {
				sts.distr.score10.notNomen++
			}
		} else if nrs[i].References[0].Score.Odds >= 1 {
			if hasNomenRef {
				sts.distr.score1.isNomen++
			} else {
				sts.distr.score1.notNomen++
			}
		} else if nrs[i].References[0].Score.Odds >= 0.1 {
			if hasNomenRef {
				sts.distr.score01.isNomen++
			} else {
				sts.distr.score01.notNomen++
			}
		} else if nrs[i].References[0].Score.Odds >= 0.01 {
			if hasNomenRef {
				sts.distr.score001.isNomen++
			} else {
				sts.distr.score001.notNomen++
			}
		} else {
			if hasNomenRef {
				sts.distr.score0.isNomen++
			} else {
				sts.distr.score0.notNomen++
			}
		}
	}
	displayStats(sts)
	output(nrs)
}

func displayStats(sts *stats) {
	fmt.Printf("Total names: %d\n", namesTotal)
	fmt.Printf("Names considered: %d\n", sts.namesNum)
	fmt.Printf("Found nomens: %d\n", sts.nameNomenNum)
	fmt.Printf("nomen/total: %0.2f%%", 100.0*float32(sts.nameNomenNum)/float32(namesTotal))
	fmt.Print("\n\nOdds distribution\n\n")
	tot, nom, p, odds := calcProb(sts.score100)
	fmt.Printf("Odds > 100.0: all %3d, nomen %3d, prob %0.3f, odds %0.3f\n", tot, nom, p, odds)
	tot, nom, p, odds = calcProb(sts.score10)
	fmt.Printf("Odds > 10.0:  all %3d, nomen %3d, prob %0.3f, odds %0.3f\n", tot, nom, p, odds)
	tot, nom, p, odds = calcProb(sts.score1)
	fmt.Printf("Odds > 1.0:   all %3d, nomen %3d, prob %0.3f, odds %0.3f\n", tot, nom, p, odds)
	tot, nom, p, odds = calcProb(sts.score01)
	fmt.Printf("Odds > 0.1:   all %3d, nomen %3d, prob %0.3f, odds %0.3f\n", tot, nom, p, odds)
	tot, nom, p, odds = calcProb(sts.score001)
	fmt.Printf("Odds > 0.01:  all %3d, nomen %3d, prob %0.3f, odds %0.3f\n", tot, nom, p, odds)
	tot, nom, p, odds = calcProb(sts.score0)
	fmt.Printf("Odds > 0:     all %3d, nomen %3d, prob %0.3f, odds %0.3f\n", tot, nom, p, odds)
}

func calcProb(d distData) (int, int, float64, float64) {
	total := d.isNomen + d.notNomen
	nomen := d.isNomen
	prob := float64(nomen) / float64(total)
	odds := prob / (1.0 - prob)
	return total, nomen, prob, odds
}

func output(nrs []*namerefs.NameRefs) {
	enc := gnfmt.GNjson{Pretty: true}
	var res []*namerefs.NameRefs
	for i := range nrs {
		var refs []*refbhl.ReferenceBHL
		for _, v := range nrs[i].References {
			if v.Score.Odds >= 0.1 {
				refs = append(refs, v)
			}
		}
		if len(refs) > 0 {
			nrs[i].References = refs
			res = append(res, nrs[i])
		}
	}
	bs, err := enc.Encode(res)
	if err != nil {
		log.Fatal().Err(err)
	}
	err = os.WriteFile(resFile, bs, 0644)
	if err != nil {
		log.Fatal().Err(err)
	}
}
