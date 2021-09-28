package score

import (
	"fmt"

	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/reffinder"
)

type Score interface {
	fmt.Stringer
	Calculate(*namerefs.NameRefs, reffinder.RefFinder) error
}
