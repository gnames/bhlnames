package score

import (
	"strconv"
	"strings"

	"github.com/gnames/bhlnames/internal/ent/refbhl"
)

func getPageScore(pageStart, pageEnd int, ref *refbhl.ReferenceNameBHL) (int, string) {
	score := 0

	if pageStart == 0 && pageEnd == 0 {
		return pagesLabel(score)
	}

	if ref.PageNum == 0 && ref.Pages == "" {
		return pagesLabel(score)
	}

	if ref.PageNum > 0 {
		if pageEnd == 0 && ref.PageNum == pageStart {
			score = 2
		} else if pageStart <= ref.PageNum && ref.PageNum <= pageEnd {
			score = 2
		}
	}

	if ref.Pages != "" {
		partPages := strings.Split(ref.Pages, "-")
		var partPageStart, partPageEnd int
		partPageStart, _ = strconv.Atoi(partPages[0])

		if partPageStart == 0 {
			return pagesLabel(score)
		}

		if len(partPages) > 1 && pageEnd > pageStart {
			partPageEnd, _ = strconv.Atoi(partPages[1])

			if pageStart >= partPageStart && pageEnd <= partPageEnd {
				score += 1
			}
		}
	}
	return pagesLabel(score)
}

func pagesLabel(score int) (int, string) {
	switch score {
	case 3:
		return score, "both"
	case 2:
		return score, "pageNum"
	case 1:
		return score, "paperPages"
	default:
		return score, "none"
	}
}
