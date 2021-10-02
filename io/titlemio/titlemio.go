package titlemio

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/aho_corasick"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/title_matcher"
	"github.com/gnames/bhlnames/io/db"
)

type titlemio struct {
	// Config contains general configuration of BHLnames. Some of the config
	// settings modify behavior of algorithms to find BHL references.
	config.Config

	// AC is AhoCorasick object for matching references to BHL titles.
	aho_corasick.AhoCorasick

	// TitleKV contains a key-value Badger store where data about titleIDs and
	// title abbreviations is kept.
	TitleKV *badger.DB
}

func New(cfg config.Config) title_matcher.TitleMatcher {
	res := &titlemio{
		Config:  cfg,
		TitleKV: db.InitKeyVal(cfg.AhoCorKeyValDir),
	}
	ac, err := res.getAhoCorasick()
	if err != nil {
		log.Fatal(err)
	}
	res.AhoCorasick = ac
	return res
}

func (tm *titlemio) Close() error {
	return tm.TitleKV.Close()
}

func (tm *titlemio) getAhoCorasick() (aho_corasick.AhoCorasick, error) {
	var err error
	var txt []byte
	var patterns []string
	ac := aho_corasick.New()

	path := filepath.Join(tm.AhoCorasickDir, "patterns.txt")
	txt, err = os.ReadFile(path)
	if err == nil {
		patterns = strings.Split(string(txt), "\n")
		for i := range patterns {
			patterns[i] = strings.TrimSpace(patterns[i])
		}
		acSize := ac.Setup(patterns)
		log.Printf("Created Title search trie with %d nodes.\n", acSize)
	}
	return ac, err
}
