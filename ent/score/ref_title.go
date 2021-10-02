package score

import (
	"github.com/gnames/bhlnames/ent/refbhl"
)

func getRefTitleScore(titleIDs map[int][]string, ref *refbhl.ReferenceBHL) int {
	var res int
	// matched abbreviations are sorted by their length
	if abbrs, ok := titleIDs[ref.TitleID]; ok {
		switch len(abbrs[0]) {
		case 10:
			res = 15
		case 9:
			res = 14
		case 8:
			res = 12
		case 7:
			res = 9
		case 6:
			res = 6
		case 5:
			res = 3
		default:
			res = 1
		}
	}
	return res
}
