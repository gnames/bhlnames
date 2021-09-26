package score

import (
	"github.com/gnames/bhlnames/ent/refbhl"
	"github.com/gnames/gnparser"
)

type annotation int

const (
	noAnnot annotation = iota
	spNov
	subsNov
	combNov
)

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

// NO_ANNOT = 15???
// SP_NOV
// f:sp v:sp = 15
// f:sp v:gen = 0
// f:sp v:ssp = 2
// f:ssp v:sp = 9
// f:ssp v:gen = 0
// f:gen v:gen = 0
// SUBSP_NOV
// f:ssp v:ssp = 15
// f:ssp v:sp = 0
// f:ssp v:gen = 0
// f:sp v:ssp = 9
// f:gen v:gen = 0
// COMB_NOV
// f:sp v:sp = 15
// f:ssp v:ssp = 15
// f:ssp v:sp = 6
// f:sp v:ssp = 2
// f:gen v:gen = 0

func getAnnotScore(ref *refbhl.ReferenceBHL) int {
	annot := NewAnnot(ref.AnnotNomen)
	cardName, cardMatchName := cardinality(ref)
	if cardName == 0 || cardMatchName == 0 {
		return 0
	}
	switch annot {
	case spNov:
		switch {
		case cardName == 2 && cardMatchName == 2:
			return 15
		case cardName == 2 && cardMatchName == 3:
			return 2
		case cardName == 3 && cardMatchName == 2:
			return 9
		default:
			return 0
		}
	case subsNov:
		switch {
		case cardName == 3 && cardMatchName == 3:
			return 15
		case cardName == 3 && cardMatchName == 2:
			return 2
		case cardName == 2 && cardMatchName == 3:
			return 6
		default:
			return 0
		}
	case combNov:
		switch {
		case cardName == 2 && cardMatchName == 2:
			return 15
		case cardName == 3 && cardMatchName == 3:
			return 15
		case cardName == 2 && cardMatchName == 3:
			return 9
		}
	case noAnnot:
		return 0
	}
	return 0
}

func cardinality(ref *refbhl.ReferenceBHL) (int32, int32) {
	cfg := gnparser.NewConfig()
	gnp := gnparser.New(cfg)
	n := gnp.ParseName(ref.Name)
	mn := gnp.ParseName(ref.MatchName)
	return int32(n.Cardinality), int32(mn.Cardinality)
}
