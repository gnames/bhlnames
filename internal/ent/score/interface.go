package score

import (
	"fmt"

	"github.com/gnames/bayes"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/bhlnames/internal/ent/title_matcher"
)

// Score interface provides methods to calculate scores that is used to
// rank matching quality of BHL metadata to provided publication references.
type Score interface {
	fmt.Stringer
	// Calculate calculates scores for a given set using heuristic and
	// machine learning methods.
	Calculate(*namerefs.NameRefs, title_matcher.TitleMatcher, bayes.Bayes) error
}
