package score

import (
	"fmt"

	"github.com/gnames/bhlnames/ent/namerefs"
)

type Score interface {
	fmt.Stringer
	Calculate(*namerefs.NameRefs)
}
