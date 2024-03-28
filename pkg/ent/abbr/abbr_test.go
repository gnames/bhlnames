package abbr_test

import (
	"testing"

	"github.com/gnames/bhlnames/internal/io/dictio"
	"github.com/gnames/bhlnames/pkg/ent/abbr"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	tests := []struct {
		msg, title string
		abbrs      []string
	}{
		{
			"1",
			"Abhandlungen der K. K. Zool.-Botan. Gesellschaft in Wien.",
			[]string{"adkkzbgiw", "adkkzbgi", "adkkzbg", "akkzbgw",
				"adkkzb", "akkzbg", "akkzb", "akkz"},
		},
		{
			"2",
			"Års-berättelse om botaniska arbeten och upptäckter ...",
			[]string{"abobaou", "abobao", "abbau", "aboba", "abba", "abob"},
		},
		{
			"3",
			"Annales du Muséum national d'histoire naturelle.",
			[]string{"admndhn", "admndh", "admnd", "amnhn", "admn", "amnh"},
		},
		{
			"4",
			"Skalitzky, C. Zwei neue europäische Staphylinenarten aus Portugal. Wiener Entomologische Zeitung, 3 (4): 97-99. (1884).",
			[]string{"scznesapwe", "scznespwez", "scznesapw", "scznespwe", "scznesap", "scznespw", "scznesa", "scznesp"},
		},
		{
			"5",
			"Bulletin de l'Herbier Boissier.",
			[]string{"bdlhb", "bdlh", "bhb"},
		},
		{
			"6",
			"Abhandlungen der K.K. Zool.-Botan.Gesellschaft in Wien.",
			[]string{"adkkzbgiw", "adkkzbgi", "adkkzbg", "akkzbgw",
				"adkkzb", "akkzbg", "akkzb", "akkz"},
		},
	}

	d := dictio.New()
	shortWords, err := d.ShortWords()
	assert.Nil(t, err)
	for _, v := range tests {
		n := abbr.Patterns(v.title, shortWords)
		assert.Equal(t, n, v.abbrs, v.msg)
	}
}

func TestAbbr(t *testing.T) {
	tests := []struct {
		msg, title, abbr, abbrMax string
	}{
		{
			"1",
			"Abhandlungen der K. K. Zool.-Botan. Gesellschaft in Wien.",
			"adkkzbgiw",
			"akkzbgw",
		},
		{
			"2",
			"Års-berättelse om botaniska arbeten och upptäckter ...",
			"abobaou",
			"abbau",
		},
		{
			"2",
			"Annales du Muséum national d'histoire naturelle.",
			"admndhn",
			"amnhn",
		},
		{"3",
			"Skalitzky, C. Zwei neue europäische Staphylinenarten aus Portugal. Wiener Entomologische Zeitung, 3 (4): 97-99. (1884).",
			"scznesapwez",
			"scznespwez",
		},
	}

	d := dictio.New()
	shortWords, err := d.ShortWords()
	assert.Nil(t, err)

	for _, v := range tests {
		n := abbr.Abbr(v.title)
		ns := abbr.AbbrMax(v.title, shortWords)

		assert.Equal(t, n, v.abbr, v.msg)
		assert.Equal(t, ns, v.abbrMax, v.msg)
	}
}

func TestDeriv(t *testing.T) {
	tests := []struct {
		abbr  string
		deriv []string
	}{
		{"adkkzbgiw", []string{"adkkzbgiw", "adkkzbgi", "adkkzbg", "adkkzb"}},
		{"adkk", []string{"adkk"}},
		{"adkkd", []string{"adkkd", "adkk"}},
	}
	for _, v := range tests {
		der := abbr.ShorterStrings(v.abbr)
		assert.Equal(t, der, v.deriv, v.abbr)
	}
}
