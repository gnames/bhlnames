package abbr_test

import (
	"testing"

	"github.com/gnames/bhlnames/ent/abbr"
	"github.com/gnames/bhlnames/io/dictio"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
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
			"2",
			"Annales du Muséum national d'histoire naturelle.",
			[]string{"admndhn", "admndh", "admnd", "amnhn", "admn", "amnh"},
		},
	}

	d := dictio.New()
	shortWords, err := d.ShortWords()
	assert.Nil(t, err)

	for _, v := range tests {
		n := abbr.All(v.title, shortWords)
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
		der := abbr.Derivatives(v.abbr)
		assert.Equal(t, der, v.deriv, v.abbr)
	}
}
