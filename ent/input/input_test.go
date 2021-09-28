package input_test

import (
	"testing"

	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/gnparser"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		msg, id, name, ref                   string
		iID, iName, iCan, iAuth, iYear, iRef string
	}{
		{
			"can. only", "", "Bubo bubo", "",
			"", "Bubo bubo", "Bubo bubo", "", "", "",
		},
		{
			"id, auth, yr", "123", "Bubo bubo (Linnaeus, 1758)", "",
			"123", "Bubo bubo (Linnaeus, 1758)", "Bubo bubo", "Linnaeus", "1758", "",
		},
		{
			"bad name", "", "1-noname", "",
			"", "1-noname", "", "", "", "",
		},
		{"ref", "234", "Achenium lusitanicum Skalitzky, 1884",
			"Skalitzky, C. Zwei neue europäische Staphylinenarten aus Portugal. " +
				"Wiener Entomologische Zeitung, 3 (4): 97-99. (1884).",
			"234", "Achenium lusitanicum Skalitzky, 1884", "Achenium lusitanicum",
			"Skalitzky", "1884",
			"Skalitzky, C. Zwei neue europäische Staphylinenarten aus Portugal. " +
				"Wiener Entomologische Zeitung, 3 (4): 97-99. (1884).",
		},
	}
	gnp := gnparser.New(gnparser.NewConfig())

	for _, v := range tests {
		opts := []input.Option{
			input.OptID(v.id),
			input.OptNameString(v.name),
			input.OptRefString(v.ref),
		}
		inp := input.New(gnp, opts...)
		if v.id == "" {
			assert.Equal(t, len(inp.ID), 36, v.msg)
		} else {
			assert.Equal(t, inp.ID, v.iID, v.msg)
		}

		assert.Equal(t, inp.NameString, v.iName, v.msg)
		assert.Equal(t, inp.Canonical, v.iCan, v.msg)
	}
}
