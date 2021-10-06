package score

import (
	"math"
	"time"

	"github.com/gnames/bhlnames/ent/refbhl"
)

var maxYearScore int = 3

// getYearScore
func getYearScore(yearInput int, ref *refbhl.ReferenceBHL) (int, string) {
	var score int
	yearPart, itemYearStart, itemYearEnd, titleYearStart, titleYearEnd := getRefYears(ref)

	if yearPart > 0 {
		score = yearNear(yearInput, yearPart)
		return yearLabel(score)
	}
	var score1, score2 int
	item := int(itemYearStart+itemYearEnd) > 0
	title := int(titleYearStart+titleYearEnd) > 0
	if item || (!item && !title) {
		score1 = yearBetween(yearInput, itemYearStart, itemYearEnd)
	}
	if title {
		score2 = yearBetween(yearInput, titleYearStart, titleYearEnd)
	}

	if score1 > score2 {
		return yearLabel(score1)
	}
	return yearLabel(score2)
}

func yearLabel(score int) (int, string) {
	switch score {
	case 3:
		return score, "exact"
	case 2:
		return score, "near"
	case 1:
		return score, "far"
	default:
		return score, "none"
	}
}

func invalidYear(year int) bool {
	return year < 1740 || year > (time.Now().Year()+2)
}

func yearNear(year1, year2 int) int {
	if invalidYear(year1) {
		return 0
	}

	coef := 0.7
	dif := math.Abs(float64(year1) - float64(year2))
	if dif > 10 {
		return 0
	}
	scoreVal := float64(maxYearScore) * math.Pow(float64(coef), dif)
	return int(math.Round(scoreVal))
}

func yearBetween(year, yearMin, yearMax int) int {
	if invalidYear(year) {
		return 0
	}
	if yearMin == 0 && yearMax == 0 {
		return 0
	}

	if yearMax < yearMin && yearMax != 0 {
		return 0
	}

	if yearMax == 0 {
		return yearNear(year, yearMin)
	}

	if !(year <= yearMax && year >= yearMin) {
		return 0
	}

	return yearNear(year, yearMax)
}

func getRefYears(ref *refbhl.ReferenceBHL) (int, int, int, int, int) {
	var yearPart int
	if ref.YearType == "Part" {
		yearPart = ref.YearAggr
	}
	return yearPart, ref.ItemYearStart, ref.ItemYearEnd, ref.TitleYearStart, ref.TitleYearEnd
}
