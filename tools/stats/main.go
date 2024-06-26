package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/bhlnames/internal/ent/refbhl"
	"github.com/gnames/bhlnames/internal/ent/score"
	"github.com/gnames/bhlnames/internal/io/bayesio"
	"github.com/gnames/bhlnames/internal/io/titlemio"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnfmt"
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
	score10, score1, score01, score001, score0 distData
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
		slog.Error("Cannot read gold file.", "error", err)
		os.Exit(1)
	}
	var nrs []*namerefs.NameRefs
	enc := gnfmt.GNjson{Pretty: true}
	err = enc.Decode(gold, &nrs)
	if err != nil {
		slog.Error("Cannot decode gold file.", "error", err)
		os.Exit(1)
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
		sc.Calculate(nrs[i], tm, nb, true)
		if nrs[i].References[0].Score.Odds >= 10 {
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
	tot, nom, p, odds := calcProb(sts.score10)
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
		slog.Error("Cannot encode results.", "error", err)
		os.Exit(1)
	}
	err = os.WriteFile(resFile, bs, 0644)
	if err != nil {
		slog.Error("Cannot write results.", "error", err)
		os.Exit(1)
	}
}
