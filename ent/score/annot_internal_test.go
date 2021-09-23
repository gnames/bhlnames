package score

import (
	"testing"

	"github.com/gnames/bhlnames/ent/refbhl"
)

func TestAnnotScore(t *testing.T) {

	type data struct {
		name       string
		matchName  string
		annotation annotation
		score      int
	}

	dataArray := []data{
		//  Name
		// 1 MatchName
		// 2 Annotation
		// 3
		{"Aus bus", "Aus bus", spNov, 15},
		{"Aus bus cus", "Aus bus cus", spNov, 0.0},
		{"Aus bus cus", "Aus bus cus", subsNov, 15},
		{"Aus bus", "Aus bus cus", spNov, 2},
		{"Aus bus cus", "Aus bus", spNov, 9},
		{"Aus bus Ower", "Bus cus Mozzherin", spNov, 15},
		{"Aus (Bus) cus", "Aus cus", spNov, 15},
		{"Aus bus", "Aus bus", subsNov, 0},
		{"Aus bus", "Aus bus cus", subsNov, 6},
		{"Aus bus cus", "Aus bus", subsNov, 2},
		{"Aus bus", "Aus bus", combNov, 15},
		{"Aus bus cus", "Aus bus cus", combNov, 15},
		{"Aus bus", "Aus bus cus", combNov, 9},
		{"Aus", "Aus", combNov, 0},
		{"Aus bus cus", "Aus", noAnnot, 0.0},
		{"Aus bus", "Aus bus", noAnnot, 0.0},
		{"Aus virus", "Bus cus", spNov, 0.0},
	}

	for _, d := range dataArray {
		testRef := refbhl.ReferenceBHL{
			Name:       d.name,
			MatchName:  d.matchName,
			AnnotNomen: d.annotation.String(),
		}
		result := getAnnotScore(&testRef)
		if result != d.score {
			t.Errorf("Wrong score for AnnotScore(%#v) %d instead of %d", testRef, result, d.score)
		}

	}
}
