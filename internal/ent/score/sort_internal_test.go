package score

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Year has higher precedence over Annotation
var prec = map[ScoreType]int{Annot: 0, Year: 1}

func TestValue(t *testing.T) {
	tests := []struct {
		msg    string
		annot  int
		yr     int
		total  int
		valInt int
		val    string
	}{
		{"a2y15", 0, 0, 0, 0, "00000000_00000000_00000000_00000000"},
		{"a2y15", 1, 1, 2, 33554449, "00000010_00000000_00000000_00010001"},
		{"a15y15", 15, 15, 30, 503316735, "00011110_00000000_00000000_11111111"},
		{"a6y11", 6, 11, 17, 285212854, "00010001_00000000_00000000_10110110"},
		{"a11y6", 11, 6, 17, 285212779, "00010001_00000000_00000000_01101011"},
	}

	for _, v := range tests {
		s := &score{
			precedence: prec,
			annot:      v.annot,
			year:       v.yr,
		}
		s.combineScores()
		val := s.String()
		assert.Equal(t, v.total, s.total)
		assert.Equal(t, v.val, val)
		assert.Equal(t, v.valInt, int(s.value))
	}
}
