package score

import (
	"fmt"

	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/refbhl"
	"github.com/gnames/bhlnames/ent/title_matcher"
)

type score struct {
	total, year, annot, refTitle, refVolume, refPages int
	value                                             uint32
	precedence                                        map[ScoreType]int
}

func New(prec map[ScoreType]int) Score {
	return &score{precedence: prec}
}

func (s *score) Calculate(nr *namerefs.NameRefs, tm title_matcher.TitleMatcher) error {
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
		s.year = getYearScore(yr, refs[i])
		s.annot = getAnnotScore(refs[i])
		s.refTitle = getRefTitleScore(titleIDs, refs[i])
		s.refVolume = getVolumeScore(nr.Input.Volume, refs[i])
		s.refPages = getPageScore(nr.Input.PageStart, nr.Input.PageEnd, refs[i])
		s.combineScores()
		refs[i].Score = refbhl.Score{
			Sort:      s.value,
			Total:     s.total,
			Annot:     s.annot,
			Year:      s.year,
			RefTitle:  s.refTitle,
			RefVolume: s.refVolume,
			RefPages:  s.refPages,
		}
	}
	return nil
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
