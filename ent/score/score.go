package score

import (
	"fmt"

	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/refbhl"
	"github.com/gnames/bhlnames/ent/reffinder"
)

type score struct {
	total      int
	year       int
	annot      int
	refTitle   int
	value      uint32
	precedence map[ScoreType]int
}

func New(prec map[ScoreType]int) Score {
	return &score{precedence: prec}
}

func (s *score) Calculate(nr *namerefs.NameRefs, rf reffinder.RefFinder) error {
	refs := nr.References
	yr := getYear(nr.Input)
	for i := range refs {
		s = &score{precedence: s.precedence}
		s.year = getYearScore(yr, refs[i])
		s.annot = getAnnotScore(refs[i])

		if nr.Input.RefString != "" {
			titleScore, err := getRefTitleScore(nr.Input.RefString, refs[i], rf)
			if err != nil {
				return err
			}
			s.refTitle = titleScore
		}
		s.combineScores()
		refs[i].Score = refbhl.Score{
			Sort:     s.value,
			Total:    s.total,
			Annot:    s.annot,
			Year:     s.year,
			RefTitle: s.refTitle,
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
	s.total = s.year + s.annot + s.refTitle
	annotShift := 4 * s.precedence[Annot]
	yearShift := 4 * s.precedence[Year]
	refTitleShift := 4 * s.precedence[RefTitle]
	totalShift := 24
	s.value = (s.value | uint32(s.annot)<<annotShift)
	s.value = (s.value | uint32(s.year)<<yearShift)
	s.value = (s.value | uint32(s.refTitle)<<refTitleShift)
	s.value = (s.value | uint32(s.total)<<totalShift)
}

func getYear(inp input.Input) string {
	if inp.RefYear != "" {
		return inp.RefYear
	}
	return inp.NameYear
}