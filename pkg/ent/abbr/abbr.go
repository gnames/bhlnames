package abbr

import (
	"cmp"
	"slices"
	"strings"
	"unicode"

	"github.com/gnames/bhlnames/internal/ent/str"
	"github.com/gnames/gner/ent/token"
)

func Patterns(s string, d map[string]struct{}) []string {
	res1 := Abbr(s)
	if len(res1) > 10 {
		res1 = res1[0:10]
	}
	res2 := AbbrMax(s, d)
	if len(res2) > 10 {
		res2 = res1[0:10]
	}
	der1 := Derivatives(res1)
	if res1 == res2 {
		return der1
	}

	der2 := Derivatives(res2)
	derMap := make(map[string]struct{})
	for i := range der1 {
		derMap[der1[i]] = struct{}{}
	}
	for i := range der2 {
		derMap[der2[i]] = struct{}{}
	}
	res := make([]string, len(derMap))
	var count int
	for k := range derMap {
		res[count] = k
		count++
	}
	slices.SortFunc(res, func(a, b string) int {
		la := len(a)
		lb := len(b)
		if la != lb {
			return cmp.Compare(lb, la)
		}
		return cmp.Compare(a, b)
	})
	return res
}

// Abbr returns the "long" abbreviation of a string. For example, "Journal of
// the Linnean Society" becomes "jotls". such short strings are used with the
// Aho-Corasick algorithm to find matches of journal titles in a reference to
// the titles in the BHL database.
func Abbr(s string) string {
	return abbr(s, nil)
}

// AbbrMax abbreviates a string ignoring common short words. For example,
// "Journal of the Linnean Society" becomes "jls". Such abbreviations are used
// with the Aho-Corasick algorithm to find matches of journal titles in a
// reference to the titles in the BHL database.
func AbbrMax(s string, shortWords map[string]struct{}) string {
	return abbr(s, shortWords)
}

// Derivatives returns shortened versions of a string.
func Derivatives(s string) []string {
	if len(s) < 5 {
		return []string{s}
	}
	res := make([]string, 0, len(s)-3)
	var count int
	res = append(res, s)
	for len(s) > 4 && count < 3 {
		s = s[:len(s)-1]
		res = append(res, s)
		count += 1
	}
	return res
}

func abbr(s string, shortWords map[string]struct{}) string {
	strs := str.SplitAny(s, " .-'")
	s = strings.Join(strs, " ")
	tokens := token.Tokenize(
		[]rune(s),
		func(t token.TokenNER) token.TokenNER { return t },
	)
	var res []byte
	for _, v := range tokens {
		word, _ := str.UtfToAscii(v.Cleaned())
		if shortWords != nil {
			if _, ok := shortWords[word]; ok {
				continue
			}
		}
		if r := firstLetter(word); r != '�' {
			res = append(res, byte(r))
		}
	}
	return string(res)
}

func firstLetter(s string) rune {
	if s == "and" {
		s = "&"
	}
	for _, v := range s {
		if (unicode.IsLetter(v) && v < unicode.MaxASCII) || v == '&' {
			return unicode.ToLower(v)
		}
	}
	return '�'
}
