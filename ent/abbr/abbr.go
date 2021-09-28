package abbr

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/gnames/bhlnames/ent/str"
	"github.com/gnames/gner/ent/token"
)

func All(s string, d map[string]struct{}) []string {
	res1 := Abbr(s)
	res2 := AbbrMax(s, d)
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
	sort.Slice(res, func(i, j int) bool {
		li := len(res[i])
		lj := len(res[j])
		if li != lj {
			return li > lj
		}
		return res[i] < res[j]
	})
	return res
}

func Abbr(s string) string {
	return abbr(s, nil)
}

func AbbrMax(s string, shortWords map[string]struct{}) string {
	return abbr(s, shortWords)
}

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
	strs := str.SplitAny(s, " -'")
	s = strings.Join(strs, " ")
	tokens := token.Tokenize(
		[]rune(s),
		func(t token.TokenNER) token.TokenNER { return t },
	)
	var res []byte
	for _, v := range tokens {
		word, _ := str.NormUTF(v.Cleaned())
		if shortWords != nil {
			if _, ok := shortWords[word]; ok {
				continue
			}
		}
		if r := firstLetter(word); r != '�' {
			res = append(res, byte(r))
		}
	}
	if len(res) > 10 {
		res = res[:10]
	}
	if !str.IsASCII(string(res)) {
		fmt.Println(string(res))
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