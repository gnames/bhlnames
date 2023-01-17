package input_test

import (
	"testing"

	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/gnparser"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		msg, id, name, ref                                                            string
		iID, iName, iCan, iAuth                                                       string
		iNameYear, iRefYearStart, iRefYearEnd, iRefVolume, iRefPageStart, iRefPageEnd int
	}{
		{
			"can. only", "", "Bubo bubo", "",
			"", "Bubo bubo", "Bubo bubo", "", 0, 0, 0, 0, 0, 0,
		},
		{
			"id, auth, yr", "123", "Bubo bubo (Linnaeus, 1758)", "",
			"123", "Bubo bubo (Linnaeus, 1758)", "Bubo bubo", "Linnaeus", 1758, 0, 0, 0, 0, 0,
		},
		{
			"bad name", "", "1-noname", "",
			"", "1-noname", "", "", 0, 0, 0, 0, 0, 0,
		},
		{
			"ref", "234", "Achenium lusitanicum Skalitzky, 1884",
			"Skalitzky, C. Zwei neue europäische Staphylinenarten aus Portugal. " +
				"Wiener Entomologische Zeitung, 3 (4): 97-99. (1884).",
			"234", "Achenium lusitanicum Skalitzky, 1884", "Achenium lusitanicum",
			"Skalitzky", 1884, 1884, 0, 3, 97, 99,
		},
		{
			"ref2", "74738", "Tasmanoonops inornatus Hickman, 1979", "Hickman, V. V. Some Tasmanian spiders of the families Oonopidae, Anapidae and Mysmenidae. Papers and Proceedings of the Royal Society of Tasmania 113: 53-79. (1979).",
			"74738", "Tasmanoonops inornatus Hickman, 1979",
			"Tasmanoonops inornatus", "Hickman", 1979,
			1979, 0, 113, 53, 79,
		},
		{
			"ref3", "38483", "Scapanes australis (Boisduval, 1835)", "Boisduval J.B. Voyage de découvertes de l’Astrolabe. Exécuté par ordre du Roi, pendant les années 1826, 1827, 1829, sous le commandement de M.J.Dumont d'Urville. Faune Entomologique de l'Océan Pacifique, avec l'illustration des insectes nouveaux receuillis pendant le voyage. 2me partie. Coléoptères et autres Ordres. Tatsu. Paris 2:1-716 (152-247). (1835).",
			"38483", "Scapanes australis (Boisduval, 1835)",
			"Scapanes australis", "Boisduval", 1835,
			1835, 0, 2, 1, 716,
		},
		{
			"ref4", "48949", "Meniscium guyanense Fée", "Fée. In: Gen. 224. (1850-52).",
			"48949", "Meniscium guyanense Fée",
			"Meniscium guyanense", "Fée", 0,
			1850, 1852, 0, 0, 0,
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

		assert.Equal(t, inp.RefYearStart, v.iRefYearStart, v.msg)
		assert.Equal(t, inp.RefYearEnd, v.iRefYearEnd, v.msg)
		assert.Equal(t, inp.Volume, v.iRefVolume, v.msg)
		assert.Equal(t, inp.PageStart, v.iRefPageStart, v.msg)
		assert.Equal(t, inp.PageEnd, v.iRefPageEnd, v.msg)
		assert.Equal(t, inp.NameString, v.iName, v.msg)
		assert.Equal(t, inp.Canonical, v.iCan, v.msg)
	}
}
