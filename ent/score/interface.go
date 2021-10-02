package score

import (
	"fmt"

	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/title_matcher"
)

type Score interface {
	fmt.Stringer
	Calculate(*namerefs.NameRefs, title_matcher.TitleMatcher) error
}
