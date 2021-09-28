package score

import (
	"math"
	"strconv"
	"testing"

	"github.com/gnames/bhlnames/ent/refbhl"
)

func TestYearNear(t *testing.T) {
	// yearNear does not check for validness of the year, it is done in
	// YearScore function.
	years := [][]int{{2001, 2000}, {2000, 2001}, {2000, 2000}, {2000, 2002}, {2000, 2003}, {-1, -1}, {3000, 3001}}
	scores := []int{11, 11, 15, 7, 5, 0, 0}
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
	type data struct {
		values []int
		score  int
	}

	dataArray := []data{
		{[]int{0, 0, 0}, 0},
		{[]int{0, 2000, 2001}, 0},
		{[]int{0, 2000, 0}, 0},
		{[]int{0, 0, 2000}, 0},
		{[]int{2000, 0, 0}, 1},
		{[]int{2000, 2000, 2001}, 11},
		{[]int{1999, 2000, 2001}, 0},
		{[]int{2002, 2000, 2001}, 0},
		{[]int{2001, 2001, 0}, 15},
		{[]int{2001, 2002, 0}, 11},
		{[]int{2001, 2003, 0}, 7},
		{[]int{2003, 2002, 0}, 11},
		{[]int{2003, 2003, 2003}, 15},
		{[]int{2002, 1993, 2003}, 11},
		{[]int{1993, 1993, 2003}, 0},
		{[]int{1981, 1980, 2003}, 0},
		{[]int{3000, 3000, 3000}, 0},
		{[]int{0, 3000, 3000}, 0},
		{[]int{3000, 0, 0}, 0},
		{[]int{0, 0, 3000}, 0},
		{[]int{0, 3000, 0}, 0},
	}

	for _, d := range dataArray {
		score := yearBetween(d.values[0], d.values[1], d.values[2])
		if math.Abs(float64(score)-float64(d.score)) > 0.0001 {
			t.Errorf("Wrong score for YearsBetween(%d, %d, %d): %d instead of %d",
				d.values[0], d.values[1], d.values[2], score, d.score)
		}
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
		{"Part", []int{2000, 2000, 2000, 0, 0}, 2000, 15},
		{"Part", []int{2000, 0, 0, 1990, 2001}, 2000, 15},
		{"Title", []int{2000, 0, 0, 2000, 0}, 2000, 15},
		{"Part", []int{0, 2000, 2001, 0, 0}, 2000, 11},
		{"Title", []int{1837, 0, 0, 1837, 1858}, 1849, 1},
		{"Title", []int{1837, 0, 0, 1837, 1858}, 1838, 0},
		{"Item", []int{1837, 1837, 1858, 0, 0}, 1849, 1},
		{"Item", []int{1837, 1837, 1858, 0, 0}, 1838, 0},
		{"Item", []int{1837, 1837, 1839, 1837, 1890}, 1838, 11},
		{"Item", []int{1837, 1837, 1849, 1837, 1838}, 1838, 15},
		{"Title", []int{0, 0, 0, 0, 0}, 1849, 1},
	}

	for _, d := range dataArray {
		testRef := refbhl.ReferenceBHL{
			YearType:       d.refType,
			YearAggr:       d.refYears[0],
			ItemYearStart:  d.refYears[1],
			ItemYearEnd:    d.refYears[2],
			TitleYearStart: d.refYears[3],
			TitleYearEnd:   d.refYears[4],
		}

		result := getYearScore(strconv.Itoa(d.year), &testRef)

		if result != d.score {
			t.Errorf("Wrong score for YearScore(%d, %#v) %d %d\n\n", d.year, testRef, result, d.score)
		}
	}
}
