package builderio

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/internal/ent/abbr"
	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/gnames/bhlnames/internal/io/dictio"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnfmt"
	"github.com/rs/zerolog/log"
)

type titleStore struct {
	cfg        config.Config
	titles     map[int]*title
	shortWords map[string]struct{}
	abbrMap    map[string][]int
}

func newTitleStore(cfg config.Config, titles map[int]*title) *titleStore {
	d := dictio.New()
	shortWords, err := d.ShortWords()
	if err != nil {
		err = fmt.Errorf("builderio.newTitleStore: %#w", err)
		log.Fatal().Err(err).Msg("newTitleStore")
	}
	res := titleStore{
		cfg:        cfg,
		titles:     titles,
		shortWords: shortWords,
	}
	return &res
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
	kv := db.InitKeyVal(ts.cfg.AhoCorKeyValDir)
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

	sort.Slice(abbrs, func(i, j int) bool {
		return len(abbrs[i]) > len(abbrs[j])
	})
	tmp := strings.Join(abbrs, "\n")
	path := filepath.Join(ts.cfg.AhoCorasickDir, "patterns.txt")
	return os.WriteFile(path, []byte(tmp), 0644)
}
