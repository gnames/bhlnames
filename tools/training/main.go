package main

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/gnames/bayes"
	ft "github.com/gnames/bayes/ent/feature"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/gnfmt"
	"github.com/rs/zerolog/log"
)

var dataPath = filepath.Join("..", "..", "io", "bayesio", "data")
var outputFile = filepath.Join(dataPath, "bayes.json")
var goldFile = filepath.Join(dataPath, "gold.json")

type label string

func (l label) String() string {
	return string(l)
}

func main() {
	gold, err := os.ReadFile(goldFile)
	if err != nil {
		log.Fatal().Err(err)
	}
	var data []*namerefs.NameRefs
	enc := gnfmt.GNjson{Pretty: true}
	err = enc.Decode(gold, &data)
	if err != nil {
		log.Fatal().Err(err)
	}
	var lfs []ft.LabeledFeatures
	for _, v := range data {
		lfs = append(lfs, bayesData(v)...)
	}
	nb := bayes.New()
	nb.Train(lfs)
	nbDump, err := nb.Dump()
	if err != nil {
		log.Fatal().Err(err)
	}
	err = os.WriteFile(outputFile, nbDump, 0644)
	if err != nil {
		log.Fatal().Err(err)
	}
}

func bayesData(nr *namerefs.NameRefs) []ft.LabeledFeatures {
	var res []ft.LabeledFeatures
	for i, v := range nr.References {
		label := ft.Label("notNomen")
		if v.IsNomenRef {
			label = ft.Label("isNomen")
		}
		bestRes := strconv.FormatBool(i == 0)
		resNum := "many"
		if nr.ReferenceNumber <= 5 {
			resNum = "few"
		}
		res = append(res, ft.LabeledFeatures{
			Label: label,
			Features: []ft.Feature{
				{Name: ft.Name("bestRes"), Value: ft.Val(bestRes)},
				{Name: ft.Name("resNum"), Value: ft.Val(resNum)},
				{Name: ft.Name("yrPage"), Value: getYearPage(v.Score.Labels)},
				{Name: ft.Name("annot"), Value: ft.Val(v.Score.Labels["annot"])},
				{Name: ft.Name("title"), Value: ft.Val(v.Score.Labels["title"])},
				{Name: ft.Name("vol"), Value: ft.Val(v.Score.Labels["vol"])},
			},
		})
	}
	return res
}

func getYearPage(ls map[string]string) ft.Val {
	page := "true"
	if ls["pages"] == "none" {
		page = "false"
	}

	l := ls["year"] + "|" + page
	return ft.Val(l)
}
