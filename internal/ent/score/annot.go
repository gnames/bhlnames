package score

import (
	"github.com/gnames/bhlnames/internal/ent/refbhl"
	"github.com/gnames/gnparser"
)

type annotation int

const (
	noAnnot annotation = iota
	spNov
	subsNov
	combNov
)

// NewAnnot coverts annotation string to annotation type (if possible).
func NewAnnot(annot string) annotation {
	annotations := map[string]annotation{
		"NO_ANNOT":  noAnnot,
		"SP_NOV":    spNov,
		"SUBSP_NOV": subsNov,
		"COMB_NOV":  combNov,
	}
	if a, ok := annotations[annot]; ok {
		return a
	}
	return noAnnot
}

// String returns string representation of annotation type.
func (a annotation) String() string {
	switch int(a) {
	case 1:
		return "SP_NOV"
	case 2:
		return "SUBSP_NOV"
	case 3:
		return "COMB_NOV"
	default:
		return "NO_ANNOT"
	}
}

// NO_ANNOT = 3???
// SP_NOV
// f:sp v:sp = 3
// f:sp v:gen = 0
// f:sp v:ssp = 1
// f:ssp v:sp = 2
// f:ssp v:gen = 0
// f:gen v:gen = 0
// SUBSP_NOV
// f:ssp v:ssp = 3
// f:ssp v:sp = 1
// f:ssp v:gen = 0
// f:sp v:ssp = 2
// f:gen v:gen = 0
// COMB_NOV
// f:sp v:sp = 3
// f:ssp v:ssp = 3
// f:ssp v:sp = 2
// f:sp v:ssp = 1
// f:gen v:gen = 0

func getAnnotScore(ref *refbhl.ReferenceNameBHL) (int, string) {
	annot := NewAnnot(ref.AnnotNomen)
	cardName, cardMatchName := cardinality(ref)
	if cardName == 0 || cardMatchName == 0 {
		return annotLabel(0)
	}
	switch annot {
	case spNov:
		switch {
		case cardName == 2 && cardMatchName == 2:
			return annotLabel(3)
		case cardName == 2 && cardMatchName == 3:
			return annotLabel(1)
		case cardName == 3 && cardMatchName == 2:
			return annotLabel(2)
		default:
			return annotLabel(0)
		}
	case subsNov:
		switch {
		case cardName == 3 && cardMatchName == 3:
			return annotLabel(3)
		case cardName == 3 && cardMatchName == 2:
			return annotLabel(1)
		case cardName == 2 && cardMatchName == 3:
			return annotLabel(2)
		default:
			return annotLabel(0)
		}
	case combNov:
		switch {
		case cardName == 2 && cardMatchName == 2:
			return annotLabel(3)
		case cardName == 3 && cardMatchName == 3:
			return annotLabel(3)
		case cardName == 3 && cardMatchName == 2:
			return annotLabel(2)
		case cardName == 2 && cardMatchName == 3:
			return annotLabel(1)
		}
	case noAnnot:
		return annotLabel(0)
	}
	return annotLabel(0)
}

func annotLabel(score int) (int, string) {
	switch score {
	case 3:
		return score, "exact"
	case 2:
		return score, "likely"
	case 1:
		return score, "doubtful"
	default:
		return score, "none"
	}
}

func cardinality(ref *refbhl.ReferenceNameBHL) (int32, int32) {
	cfg := gnparser.NewConfig()
	gnp := gnparser.New(cfg)
	n := gnp.ParseName(ref.Name)
	mn := gnp.ParseName(ref.MatchedName)
	return int32(n.Cardinality), int32(mn.Cardinality)
}
