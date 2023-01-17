package score

import (
	"testing"

	"github.com/gnames/bhlnames/internal/ent/refbhl"
	"github.com/stretchr/testify/assert"
)

func TestAnnotScore(t *testing.T) {

	tests := []struct {
		msg        string
		name       string
		matchName  string
		annotation annotation
		score      int
	}{
		//  Name
		// 1 MatchName
		// 2 Annotation
		// 3
		{"1", "Aus bus", "Aus bus", spNov, 3},
		{"2", "Aus bus cus", "Aus bus cus", spNov, 0},
		{"3", "Aus bus cus", "Aus bus cus", subsNov, 3},
		{"4", "Aus bus", "Aus bus cus", spNov, 1},
		{"5", "Aus bus cus", "Aus bus", spNov, 2},
		{"6", "Aus bus Ower", "Bus cus Mozzherin", spNov, 3},
		{"7", "Aus (Bus) cus", "Aus cus", spNov, 3},
		{"8", "Aus bus", "Aus bus", subsNov, 0},
		{"9", "Aus bus", "Aus bus cus", subsNov, 2},
		{"10", "Aus bus cus", "Aus bus", subsNov, 1},
		{"11", "Aus bus", "Aus bus", combNov, 3},
		{"12", "Aus bus cus", "Aus bus cus", combNov, 3},
		{"13", "Aus bus", "Aus bus cus", combNov, 1},
		{"14", "Aus", "Aus", combNov, 0},
		{"15", "Aus bus cus", "Aus", noAnnot, 0},
		{"16", "Aus bus", "Aus bus", noAnnot, 0},
		{"17", "Aus virus", "Bus cus", spNov, 0},
	}

	for _, v := range tests {
		testRef := refbhl.ReferenceBHL{
			Name:       v.name,
			MatchName:  v.matchName,
			AnnotNomen: v.annotation.String(),
		}
		score, _ := getAnnotScore(&testRef)
		assert.Equal(t, score, v.score, v.msg)
	}
}
