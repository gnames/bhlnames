package ttlmchio

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gnames/bhlnames/pkg/ent/abbr"
)

func (tm *ttlmchio) TitlesBHL(refString string) (map[int][]string, error) {
	refAbbr := abbr.Abbr(refString)
	matches := tm.Search(refAbbr)

	abbrs := make([]string, len(matches))
	for i := range matches {
		abbrs[i] = matches[i].Pattern
	}
	return tm.abbrsToTitleIDs(abbrs)
}

func (tm *ttlmchio) abbrsToTitleIDs(abbrs []string) (map[int][]string, error) {
	res := make(map[int][]string)
	q := `
SELECT  DISTINCT i.title_id, title_name
  FROM abbr_titles attl
    JOIN items i
      ON i.title_id = attl.title_id
	WHERE attl.abbr = ANY($1)
`
	abbrMap := make(map[string]struct{})
	for _, abbr := range abbrs {
		abbrMap[abbr] = struct{}{}
	}

	rows, err := tm.db.Query(context.Background(), q, abbrs)
	if err != nil {
		slog.Error("Cannot get titles from abbreviations", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		err := rows.Scan(&id, &name)
		if err != nil {
			slog.Error("Cannot scan title from abbreviation", "error", err)
			return nil, err
		}
		abbrStr := nameToAbbr(name, abbrMap, tm.shortWords)
		if len(abbrStr) > 0 {
			res[id] = abbrStr
		} else {
			err := errors.New("title not found")
			slog.Error("Title not found for abbreviation", "abbr", name, "err", err)
			return nil, err
		}
	}

	return res, nil
}

func nameToAbbr(
	name string,
	abbrMap map[string]struct{},
	shortWords map[string]struct{},
) []string {
	var res []string

	abbrs := abbr.Patterns(name, shortWords)
	for i := range abbrs {
		if _, ok := abbrMap[abbrs[i]]; ok {
			res = append(res, abbrs[i])
		}
	}
	return res
}
