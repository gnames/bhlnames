package score

import (
	"testing"

	"github.com/gnames/bhlnames/ent/refbhl"
	"github.com/stretchr/testify/assert"
)

func TestPageScore(t *testing.T) {

	tests := []struct {
		msg                string
		pageStart, pageEnd int
		pageNum            int
		partPages          string
		score              int
	}{
		{"ideal", 50, 70, 55, "50-70", 15},
		{"worst", 0, 0, 0, "", 0},
		{"no part pages", 50, 70, 55, "", 12},
		{"bad pageNum, no part pages", 50, 70, 20, "", 0},
		{"missing pageEnd", 50, 0, 50, "", 12},
		{"missing input pages", 0, 0, 50, "50-70", 0},
		{"no pageNum", 66, 68, 0, "61-70", 3},
		{"bad partPages", 60, 90, 70, "nonsense", 12},
		{"no partPageStart", 60, 66, 62, "-66", 12},
		{"pageNum is outside input", 60, 66, 1, "55-69", 3},
		{"bad partPageEnd", 60, 66, 1, "55-69?", 0},
		{"input outside partPages", 60, 80, 1, "55-69", 0},
		{"exact match partPages", 55, 69, 1, "55-69", 3},
		{"out of range", 121, 0, 546, "53-85", 0},
	}

	for _, d := range tests {
		testRef := refbhl.ReferenceBHL{
			PageNum:   d.pageNum,
			PartPages: d.partPages,
		}

		assert.Equal(t, d.score, getPageScore(d.pageStart, d.pageEnd, &testRef), d.msg)
	}

}
