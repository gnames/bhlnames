package builderio

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/abbr"
	"github.com/gnames/bhlnames/ent/titlefinder"
	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/bhlnames/io/dictio"
	"github.com/gnames/gnfmt"
)

type tfio struct {
	cfg        config.Config
	titles     map[int]*title
	shortWords map[string]struct{}
	abbrMap    map[string][]int
}

func NewTitleFinder(cfg config.Config, titles map[int]*title) titlefinder.TitleFinder {
	d := dictio.New()
	shortWords, err := d.ShortWords()
	if err != nil {
		log.Fatal(err)
	}
	res := tfio{
		cfg:        cfg,
		titles:     titles,
		shortWords: shortWords,
	}
	return &res
}

func (tf *tfio) Setup() error {
	abbrMap := make(map[string][]int)
	for k, v := range tf.titles {
		abbrs := abbr.All(v.Name, tf.shortWords)
		for i := range abbrs {
			if len(abbrs[i]) > 3 {
				abbrMap[abbrs[i]] = append(abbrMap[abbrs[i]], k)
			}
		}
	}
	return tf.save(abbrMap)
}

func (tf *tfio) Search(ref string) (titleIDs map[string]int, err error) {
	return nil, nil
}

func (tf *tfio) save(abbrMap map[string][]int) error {
	var err error
	kv := db.InitKeyVal(tf.cfg.AhoCorKeyValDir)
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
	path := filepath.Join(tf.cfg.AhoCorasickDir, "patterns.txt")
	return os.WriteFile(path, []byte(tmp), 0644)
}
