package score

import (
	"testing"

	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/stretchr/testify/assert"
)

func TestYearNear(t *testing.T) {
	// yearNear does not check for validness of the year, it is done in
	// YearScore function.
	years := [][]int{
		{2001, 2000},
		{2000, 2001},
		{2000, 2000},
		{2000, 2002},
		{2000, 2003},
		{-1, -1},
		{3000, 3001},
	}
	scores := []int{2, 2, 3, 1, 1, 0, 0}
	for i, v := range years {
		score := yearNear(v[0], v[1])
		if score != scores[i] {
			t.Errorf(
				"Wrong score for YearNear(%d, %d): %d instead of %d",
				v[0],
				v[1],
				score, scores[i])
		}
	}

}

func TestYearBetween(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg    string
		values []int
		score  int
	}{
		{"1", []int{0, 0, 0}, 0},
		{"2", []int{0, 2000, 2001}, 0},
		{"3", []int{0, 2000, 0}, 0},
		{"4", []int{0, 0, 2000}, 0},
		{"5", []int{2000, 0, 0}, 0},
		{"6", []int{2000, 2000, 2001}, 2},
		{"7", []int{1999, 2000, 2001}, 0},
		{"8", []int{2002, 2000, 2001}, 0},
		{"9", []int{2001, 2001, 0}, 3},
		{"10", []int{2001, 2002, 0}, 2},
		{"11", []int{2001, 2003, 0}, 1},
		{"12", []int{2003, 2002, 0}, 2},
		{"13", []int{2003, 2003, 2003}, 3},
		{"14", []int{2002, 1993, 2003}, 2},
		{"15", []int{1993, 1993, 2003}, 0},
		{"16", []int{1981, 1980, 2003}, 0},
		{"17", []int{3000, 3000, 3000}, 0},
		{"18", []int{0, 3000, 3000}, 0},
		{"19", []int{3000, 0, 0}, 0},
		{"20", []int{0, 0, 3000}, 0},
		{"21", []int{0, 3000, 0}, 0},
	}

	for _, v := range tests {
		score := yearBetween(v.values[0], v.values[1], v.values[2])
		assert.Equal(score, v.score, v.msg)
	}
}

func TestYearScore(t *testing.T) {

	type data struct {
		refType  string
		refYears []int
		year     int
		score    int
	}

	dataArray := []data{
		// 0 YearAggr
		// 1 ItemYearStart
		// 2 ItemYearEnd
		// 3 TitleYearStart
		// 4 TitleYearEnd
		// 5 Score
		{"Part", []int{0, 0, 0, 0, 0}, 0, 0},
		{"Part", []int{0, 2000, 2001, 0, 0}, 0, 0},
		{"Part", []int{0, 2000, 2001, 0, 0}, 3000, 0},
		{"Part", []int{3000, 3000, 2001, 0, 0}, 3000, 0},
		{"Part", []int{2000, 2000, 2000, 0, 0}, 2000, 3},
		{"Part", []int{2000, 0, 0, 1990, 2001}, 2000, 3},
		{"Title", []int{2000, 0, 0, 2000, 0}, 2000, 3},
		{"Part", []int{0, 2000, 2001, 0, 0}, 2000, 2},
		{"Title", []int{1837, 0, 0, 1837, 1858}, 1849, 0},
		{"Title", []int{1837, 0, 0, 1837, 1858}, 1838, 0},
		{"Item", []int{1837, 1837, 1858, 0, 0}, 1849, 0},
		{"Item", []int{1837, 1837, 1858, 0, 0}, 1838, 0},
		{"Item", []int{1837, 1837, 1839, 1837, 1890}, 1838, 2},
		{"Item", []int{1837, 1837, 1849, 1837, 1838}, 1838, 3},
		{"Title", []int{0, 0, 0, 0, 0}, 1849, 0},
	}

	for _, d := range dataArray {
		testRef := bhl.ReferenceName{
			Reference: bhl.Reference{
				YearType:       d.refType,
				YearAggr:       d.refYears[0],
				ItemYearStart:  d.refYears[1],
				ItemYearEnd:    d.refYears[2],
				TitleYearStart: d.refYears[3],
				TitleYearEnd:   d.refYears[4],
			},
		}

		result, _ := getYearScore(d.year, &testRef)

		if result != d.score {
			t.Errorf(
				"Wrong score for YearScore(%d, %#v) %d %d\n\n",
				d.year,
				testRef,
				result,
				d.score,
			)
		}
	}
}
