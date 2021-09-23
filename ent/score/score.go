package score

import (
	"fmt"
	"strconv"

	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/refbhl"
)

type score struct {
	total      int
	year       int
	annot      int
	ref        int
	value      uint32
	precedence map[ScoreType]int
}

func New(prec map[ScoreType]int) Score {
	return &score{precedence: prec}
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

func (s *score) Calculate(nr *namerefs.NameRefs) {
	refs := nr.References
	yr := getYear(nr.Input)
	for i := range refs {
		s.year = matchYear(yr, refs[i])
		s.annot = matchAnnot(refs[i])
		refs[i].Score = refbhl.Score{
			Sort:  s.CombineScores(),
			Total: s.total,
			Annot: s.annot,
			Year:  s.year,
		}
	}
}

func (s *score) CombineScores() uint32 {
	s.total = s.year + s.annot
	annotShift := 4 * s.precedence[Annot]
	yearShift := 4 * s.precedence[Year]
	totalShift := 24
	s.value = (s.value | uint32(s.annot)<<annotShift)
	s.value = (s.value | uint32(s.year)<<yearShift)
	s.value = (s.value | uint32(s.total)<<totalShift)
	return s.value
}

func (s *score) Total() int {
	return s.total
}

func (s *score) Annot() int {
	return s.annot
}

func (s *score) Year() int {
	return s.year
}

func (s *score) SortVal() int {
	return int(s.value)
}

func getYear(inp input.Input) string {
	if inp.RefYear != "" {
		return inp.RefYear
	}
	return inp.NameYear
}

func matchYear(refYr string, ref *refbhl.ReferenceBHL) (yrScore int) {
	yr, err := strconv.Atoi(refYr)
	if err != nil {
		yr = 0
	}
	return getYearScore(yr, ref)
}

func matchAnnot(ref *refbhl.ReferenceBHL) int {
	return getAnnotScore(ref)
}
