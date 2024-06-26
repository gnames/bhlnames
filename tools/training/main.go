package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gnames/bayes"
	ft "github.com/gnames/bayes/ent/feature"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/gnfmt"
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
		slog.Error("Cannot read gold file.", "error", err)
		os.Exit(1)
	}
	var data []*namerefs.NameRefs
	enc := gnfmt.GNjson{Pretty: true}
	err = enc.Decode(gold, &data)
	if err != nil {
		slog.Error("Cannot decode gold file.", "error", err)
		os.Exit(1)
	}
	var lfs []ft.ClassFeatures
	for _, v := range data {
		lfs = append(lfs, bayesData(v)...)
	}
	nb := bayes.New()
	nb.Train(lfs)
	nbDump, err := nb.Dump()
	if err != nil {
		slog.Error("Cannot dump bayes.", "error", err)
		os.Exit(1)
	}
	err = os.WriteFile(outputFile, nbDump, 0644)
	if err != nil {
		slog.Error("Cannot write bayes.", "error", err)
		os.Exit(1)
	}
}

func bayesData(nr *namerefs.NameRefs) []ft.ClassFeatures {
	var res []ft.ClassFeatures
	for i, v := range nr.References {
		class := ft.Class("notNomen")
		if v.IsNomenRef {
			class = ft.Class("isNomen")
		}
		bestRes := strconv.FormatBool(i == 0)
		resNum := "many"
		if nr.ReferenceNumber <= 5 {
			resNum = "few"
		}
		res = append(res, ft.ClassFeatures{
			Class: class,
			Features: []ft.Feature{
				{Name: ft.Name("bestRes"), Value: ft.Value(bestRes)},
				{Name: ft.Name("resNum"), Value: ft.Value(resNum)},
				{Name: ft.Name("yrPage"), Value: getYearPage(v.Score.Labels)},
				{Name: ft.Name("annot"), Value: ft.Value(v.Score.Labels["annot"])},
				{Name: ft.Name("title"), Value: ft.Value(v.Score.Labels["title"])},
				{Name: ft.Name("vol"), Value: ft.Value(v.Score.Labels["vol"])},
			},
		})
	}
	return res
}

func getYearPage(ls map[string]string) ft.Value {
	page := "true"
	if ls["pages"] == "none" {
		page = "false"
	}

	l := ls["year"] + "|" + page
	return ft.Value(l)
}
