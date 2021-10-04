package input

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePages(t *testing.T) {

	tests := []struct {
		ref   string
		pages []int
	}{
		{
			"Courtec. & P. Roux. In: Docums Mycol. 34(nos 135-136):50. (2008).", []int{50, 0},
		}, {
			"Courtec. & P. Roux. In: Docums Mycol. 34(nos 135-136):50-51. (2008).", []int{50, 51},
		},
		{
			"Courtec. & P. Roux. In: Docums Mycol. 34:50-51. (2008).", []int{50, 51},
		},
		{
			"Courtec. & P. Roux. In: Docums Mycol. 34: P51. (2008).", []int{51, 0},
		},

		{
			"Courtec. & P. Roux. In: Docums Mycol. 34.", []int{0, 0},
		},
	}

	for _, v := range tests {

		page := parsePages(v.ref)
		assert.Equal(t, page, v.pages)
	}
}

func TestParseVolume(t *testing.T) {

	tests := []struct {
		msg, ref string
		volume   int
	}{
		{
			"1", "Courtec. & P. Roux. In: Docums Mycol. 34(nos 135-136):50. (2008).", 34,
		}, {
			"2", "Courtec. & P. Roux. In: Docums Mycol. 34(nos 135-136):50-51. (2008).", 34,
		},
		{
			"3", "Courtec. & P. Roux. In: Docums Mycol. 34:50-51. (2008).", 34,
		},
		{
			"4", "Courtec. & P. Roux. In: Docums Mycol. 34: P51. (2008).", 0,
		},
		{
			"5", "Courtec. & P. Roux. In: Docums Mycol. V: 51", 0,
		},
	}

	for _, v := range tests {

		volume := parseVolume(v.ref)
		assert.Equal(t, volume, v.volume, v.msg)
	}
}

func TestParseYears(t *testing.T) {

	tests := []struct {
		msg, ref string
		years    []int
	}{
		{
			"1", "Courtec. & P. Roux. In: Docums Mycol. 34(nos 135-136):50. (2008).", []int{2008, 0},
		},
		{
			"2", "Courtec. & P. Roux. In: Docums Mycol. 34(nos 135-136):50-51. (2008-2009).", []int{2008, 2009},
		},
		{
			"3", "Courtec. & P. Roux. In: Docums Mycol. 34:50-51. (2008-09).", []int{2008, 2009},
		},
		{
			"4", "Courtec. & P. Roux. In: Docums Mycol. 34: P51. 2008.", []int{2008, 0},
		},

		{
			"5", "Courtec. & P. Roux. In: Docums Mycol. 34.", []int{0, 0},
		},
	}

	for _, v := range tests {

		years := parseYears(v.ref)
		assert.Equal(t, years, v.years, v.msg)
	}
}
