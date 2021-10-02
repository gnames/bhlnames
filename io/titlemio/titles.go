package titlemio

import (
	"github.com/gnames/bhlnames/ent/abbr"
	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/gnfmt"
)

func (tm *titlemio) TitlesBHL(refString string) (map[int][]string, error) {
	refAbbr := abbr.Abbr(refString)
	matches := tm.Search(refAbbr)

	abbrs := make([]string, len(matches))
	for i := range matches {
		abbrs[i] = matches[i].Pattern
	}
	return tm.abbrsToTitleIDs(abbrs)
}

func (tm *titlemio) abbrsToTitleIDs(abbrs []string) (map[int][]string, error) {
	res := make(map[int][]string)
	enc := gnfmt.GNgob{}

	vals, err := db.GetValues(tm.TitleKV, abbrs)
	if err != nil {
		return res, err
	}

	for k, v := range vals {
		var ids []int
		err = enc.Decode(v, &ids)
		if err != nil {
			return res, err
		}
		for i := range ids {
			if err != nil {
				return res, err
			}
			res[ids[i]] = append(res[ids[i]], k)
		}
	}
	return res, nil
}
