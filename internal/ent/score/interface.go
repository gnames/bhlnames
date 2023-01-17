package score

import (
	"fmt"

	"github.com/gnames/bayes"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/bhlnames/internal/ent/title_matcher"
)

type Score interface {
	fmt.Stringer
	Calculate(*namerefs.NameRefs, title_matcher.TitleMatcher, bayes.Bayes) error
}
