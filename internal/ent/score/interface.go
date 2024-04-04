package score

import (
	"fmt"

	"github.com/gnames/bayes"
	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/ttlmch"
)

// Score interface provides methods to calculate scores that is used to
// rank matching quality of BHL metadata to provided publication references.
type Score interface {
	fmt.Stringer
	// Calculate calculates scores for a given set using heuristic and
	// machine learning methods.
	Calculate(*bhl.RefsByName, ttlmch.TitleMatcher, bayes.Bayes, bool) error
}
