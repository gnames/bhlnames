package builderio

import (
	"cmp"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/gnames/bhlnames/internal/io/dictio"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/bhlnames/pkg/ent/abbr"
	"github.com/gnames/gnfmt"
)

type titleStore struct {
	cfg        config.Config
	titles     map[int]*title
	shortWords map[string]struct{}
	abbrMap    map[string][]int
}

func (b *builderio) dbTitlesMap() (map[int]*title, error) {
	res := make(map[int]*title)
	var err error
	var rows *sql.Rows
	rows, err = b.DB.Query(`
SELECT
	title_id, title_name, title_year_start, title_year_end,
	title_lang, title_doi
FROM items
`)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		t := title{}
		err = rows.Scan(&t.ID, &t.Name, &t.YearStart, &t.YearEnd, &t.Language, &t.DOI)
		if err != nil {
			return res, err
		}
		res[t.ID] = &t
	}
	return res, rows.Err()
}

func newTitleStore(cfg config.Config, titles map[int]*title) (*titleStore, error) {
	d := dictio.New()
	shortWords, err := d.ShortWords()
	if err != nil {
		err = fmt.Errorf("builderio.newTitleStore: %#w", err)
		slog.Error("Cannot generate short words", "error", err)
	}
	res := titleStore{
		cfg:        cfg,
		titles:     titles,
		shortWords: shortWords,
	}
	return &res, nil
}

func (ts *titleStore) setup() error {
	abbrMap := make(map[string][]int)
	for k, v := range ts.titles {
		abbrs := abbr.Patterns(v.Name, ts.shortWords)
		for i := range abbrs {
			if len(abbrs[i]) > 2 {
				abbrMap[abbrs[i]] = append(abbrMap[abbrs[i]], k)
			}
		}
	}
	return ts.save(abbrMap)
}

func (ts *titleStore) save(abbrMap map[string][]int) error {
	var err error
	kv, err := db.InitKeyVal(ts.cfg.AhoCorKeyValDir)
	if err != nil {
		return err
	}
	defer kv.Close()

	kvTxn := kv.NewTransaction(true)
	enc := gnfmt.GNgob{}

	for k, v := range abbrMap {
		var bs []byte
		bs, err = enc.Encode(v)
		if err != nil {
			return err
		}
		if err = kvTxn.Set([]byte(k), bs); err == badger.ErrTxnTooBig {
			err = kvTxn.Commit()
			if err != nil {
				return err
			}

			kvTxn = kv.NewTransaction(true)
			err = kvTxn.Set([]byte(k), bs)
			if err != nil {
				return err
			}
		}
	}
	err = kvTxn.Commit()
	if err != nil {
		return err
	}

	abbrs := make([]string, len(abbrMap))
	var count int
	for k := range abbrMap {
		abbrs[count] = k
		count += 1
	}

	slices.SortFunc(abbrs, func(a, b string) int {
		return cmp.Compare(len(b), len(a))
	})
	tmp := strings.Join(abbrs, "\n")
	path := filepath.Join(ts.cfg.AhoCorasickDir, "patterns.txt")
	return os.WriteFile(path, []byte(tmp), 0644)
}
