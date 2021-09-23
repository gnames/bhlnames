package score

import (
	"fmt"

	"github.com/gnames/bhlnames/ent/namerefs"
)

type Score interface {
	fmt.Stringer
	Calculate(nameRefs *namerefs.NameRefs)
	CombineScores() uint32
	Total() int
	Annot() int
	Year() int
	SortVal() int
}
