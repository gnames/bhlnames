package score

import (
	"testing"

	"github.com/gnames/bhlnames/ent/refbhl"
	"github.com/stretchr/testify/assert"
)

func TestVolumeScore(t *testing.T) {

	tests := []struct {
		msg       string
		volume    int
		volumeBhl string
		score     int
	}{
		{"exact match", 87, "87", 1},
		{"no match", 87, "50", 0},
		{"matches beginning", 87, "87 (1888)", 1},
		{"matches end", 87, "vol. 87", 1},
		{"matches middle", 87, "vol.87? (1888)", 1},
		{"part of number1", 87, "vol. 287 (1888)", 0},
		{"part of number2", 87, "vol. 876 (1888)", 0},
		{"part of number3", 87, "vol. 5876 (1888)", 0},
		{"range", 87, "vol. 87-89 (1888)", 1},
	}

	for _, d := range tests {
		testRef := refbhl.ReferenceBHL{
			Volume: d.volumeBhl,
		}

		assert.Equal(t, d.score, getVolumeScore(d.volume, &testRef), d.msg)
	}
}
