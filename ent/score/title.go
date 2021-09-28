package score

import (
	"fmt"

	"github.com/gnames/aho_corasick"
	"github.com/gnames/bhlnames/ent/abbr"
	"github.com/gnames/bhlnames/ent/refbhl"
)

var maxTitleScore int = 15

func getTitleScore(refString string, ref *refbhl.ReferenceBHL, ac aho_corasick.AhoCorasick) int {
	haystack := abbr.Abbr(refString)
	matches := ac.Search(haystack)
	fmt.Printf("matches %d", len(matches))
	patterns := make([]string, len(matches))
	for i := range matches {
		patterns[i] = matches[i].Pattern
	}
	fmt.Println("%%%%%%%%%%%%%")
	fmt.Println(refString)
	fmt.Println(haystack)
	for i := range patterns {
		fmt.Println(patterns[i])
	}
	fmt.Println("%%%%%%%%%%%%%")
	return 0
}
