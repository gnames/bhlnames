package score

import (
	"strconv"
	"strings"

	"github.com/gnames/bhlnames/ent/refbhl"
)

func getPageScore(pageStart, pageEnd int, ref *refbhl.ReferenceBHL) int {
	score := 0

	if pageStart == 0 && pageEnd == 0 {
		return score
	}

	if ref.PageNum == 0 && ref.PartPages == "" {
		return score
	}

	if ref.PageNum > 0 {
		if pageEnd == 0 && ref.PageNum == pageStart {
			score = 12
		} else if pageStart <= ref.PageNum && ref.PageNum <= pageEnd {
			score = 12
		}
	}

	if ref.PartPages != "" {
		partPages := strings.Split(ref.PartPages, "-")
		var partPageStart, partPageEnd int
		partPageStart, _ = strconv.Atoi(partPages[0])

		if partPageStart == 0 {
			return score
		}

		if len(partPages) > 1 && pageEnd > pageStart {
			partPageEnd, _ = strconv.Atoi(partPages[1])

			if pageStart >= partPageStart && pageEnd <= partPageEnd {
				score += 3
			}
		}
	}
	return score
}
