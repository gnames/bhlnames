package score

import (
	"fmt"

	"github.com/gnames/bayes"
	ft "github.com/gnames/bayes/ent/feature"
	bout "github.com/gnames/bayes/ent/output"
	"github.com/gnames/bayes/ent/posterior"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/refbhl"
	"github.com/gnames/bhlnames/ent/title_matcher"
	"github.com/rs/zerolog/log"
)

type score struct {
	odds                                              float64
	total, year, annot, refTitle, refVolume, refPages int
	yearLabel, annotLabel, titleLabel, volLabel       string
	pagesLabel, resNumLabel                           string
	value                                             uint32
	precedence                                        map[ScoreType]int
}

func New(prec map[ScoreType]int) Score {
	return &score{precedence: prec}
}

func (s *score) Calculate(
	nr *namerefs.NameRefs,
	tm title_matcher.TitleMatcher,
	nb bayes.Bayes,
) error {
	var err error
	refs := nr.References
	yr := getYear(nr.Input)

	refString := nr.Input.RefString
	var titleIDs map[int][]string
	if refString != "" {
		titleIDs, err = tm.TitlesBHL(refString)
		if err != nil {
			return err
		}
	}

	for i := range refs {
		s = &score{precedence: s.precedence}
		s.year, s.yearLabel = getYearScore(yr, refs[i])
		s.annot, s.annotLabel = getAnnotScore(refs[i])
		s.refTitle, s.titleLabel = getRefTitleScore(titleIDs, refs[i])
		s.refVolume, s.volLabel = getVolumeScore(nr.Input.Volume, refs[i])
		s.refPages, s.pagesLabel = getPageScore(nr.Input.PageStart, nr.Input.PageEnd, refs[i])
		s.combineScores()
		s.resNumLabel = "many"
		if nr.ReferenceNumber <= 5 {
			s.resNumLabel = "few"
		}
		postOdds, _ := s.calculateOdds(nb)
		oddsVal := postOdds.ClassOdds[ft.Class("isNomen")]
		detail := bout.New(postOdds, "isNomen")
		refs[i].Score = refbhl.Score{
			Odds:       oddsVal,
			OddsDetail: detail,
			Total:      s.total,
			Annot:      s.annot,
			Year:       s.year,
			RefTitle:   s.refTitle,
			RefVolume:  s.refVolume,
			RefPages:   s.refPages,
			Labels: map[string]string{
				"year":  s.yearLabel,
				"annot": s.annotLabel,
				"title": s.titleLabel,
				"vol":   s.volLabel,
				"pages": s.pagesLabel,
			},
		}
	}
	return nil
}

func (s *score) calculateOdds(nb bayes.Bayes) (posterior.Odds, error) {
	lfs := []ft.Feature{
		{Name: ft.Name("yrPage"), Value: s.getYearPage()},
		{Name: ft.Name("annot"), Value: ft.Value(s.annotLabel)},
		{Name: ft.Name("title"), Value: ft.Value(s.titleLabel)},
		{Name: ft.Name("vol"), Value: ft.Value(s.volLabel)},
		{Name: ft.Name("pages"), Value: ft.Value(s.pagesLabel)},
		{Name: ft.Name("resNum"), Value: ft.Value(s.resNumLabel)},
	}
	return nb.PosteriorOdds(lfs)
}

func (s *score) getYearPage() ft.Value {
	page := "true"
	if s.pagesLabel == "none" {
		page = "false"
	}

	l := s.yearLabel + "|" + page
	return ft.Value(l)
}

func (s *score) String() string {
	str := fmt.Sprintf("%032b", s.value)
	res := make([]byte, 35)
	offset := 0
	for i, v := range []byte(str) {
		res[i+offset] = v
		if (i+1)%8 == 0 && (i+1)%32 != 0 {
			offset++
			res[i+offset] = '_'
		}
	}
	return string(res)
}

// BoostBestResult provides additional score for the best result in
// NameRefs.
func BoostBestResult(nr *namerefs.NameRefs, nb bayes.Bayes) {
	if len(nr.References) > 0 {
		f := ft.Feature{Name: ft.Name("bestRes"), Value: ft.Value("true")}
		bestRes, err := nb.Likelihood(f, ft.Class("isNomen"))
		if err != nil {
			log.Fatal().Err(err)
		}
		nr.References[0].Score.Odds *= bestRes
	}
}

func (s *score) combineScores() {
	s.total = s.year + s.annot + s.refTitle + s.refVolume + s.refPages
	annotShift := 4 * s.precedence[Annot]
	yearShift := 4 * s.precedence[Year]
	refTitleShift := 4 * s.precedence[RefTitle]
	refVolume := 4 * s.precedence[RefVolume]
	refPages := 4 * s.precedence[RefPages]
	totalShift := 24
	s.value = (s.value | uint32(s.annot)<<annotShift)
	s.value = (s.value | uint32(s.year)<<yearShift)
	s.value = (s.value | uint32(s.refTitle)<<refTitleShift)
	s.value = (s.value | uint32(s.refVolume)<<refVolume)
	s.value = (s.value | uint32(s.refPages)<<refPages)
	s.value = (s.value | uint32(s.total)<<totalShift)
}

func getYear(inp input.Input) int {
	if inp.RefYearStart != 0 {
		return inp.RefYearStart
	}
	return inp.NameYear
}
