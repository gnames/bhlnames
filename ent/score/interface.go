package score

import (
	"fmt"
)

type Score interface {
	fmt.Stringer
	CombineScores()

	SetTotal(int)
	SetYear(int)
	SetAnnot(int)
	SetSortVal(uint32)

	Total() int
	Annot() int
	Year() int
	SortVal() uint32
}
