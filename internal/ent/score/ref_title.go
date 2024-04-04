package score

import "github.com/gnames/bhlnames/internal/ent/bhl"

func getRefTitleScore(
	titleIDs map[int][]string,
	ref *bhl.ReferenceName,
) (int, string) {
	var score int
	// matched abbreviations are sorted by their length
	if abbrs, ok := titleIDs[ref.TitleID]; ok {
		switch len(abbrs[0]) {
		case 10:
			score = 3
		case 9:
			score = 3
		case 8:
			score = 3
		case 7:
			score = 2
		case 6:
			score = 2
		case 5:
			score = 2
		default:
			score = 1
		}
	}
	return titleLabel(score)
}

func titleLabel(score int) (int, string) {
	switch score {
	case 3:
		return score, "long"
	case 2:
		return score, "medium"
	case 1:
		return score, "short"
	default:
		return score, "none"
	}
}
