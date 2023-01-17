package titlemio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/aho_corasick"
	"github.com/gnames/bhlnames/internal/ent/title_matcher"
	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/rs/zerolog/log"
)

type titlemio struct {
	// acDir is the directory for AhoCorasick files.
	acDir string

	// AC is AhoCorasick object for matching references to BHL titles.
	aho_corasick.AhoCorasick

	// TitleKV contains a key-value Badger store where data about titleIDs and
	// title abbreviations is kept.
	TitleKV *badger.DB
}

func New(cfg config.Config) title_matcher.TitleMatcher {
	res := &titlemio{
		acDir:   cfg.AhoCorasickDir,
		TitleKV: db.InitKeyVal(cfg.AhoCorKeyValDir),
	}
	ac, err := res.getAhoCorasick()
	if err != nil {
		err = fmt.Errorf("titlemio.New: %w", err)
		log.Fatal().Err(err).Msg("")
	}
	res.AhoCorasick = ac
	return res
}

func (tm titlemio) Close() error {
	return tm.TitleKV.Close()
}

func (tm titlemio) getAhoCorasick() (aho_corasick.AhoCorasick, error) {
	var err error
	var txt []byte
	var patterns []string
	ac := aho_corasick.New()

	path := filepath.Join(tm.acDir, "patterns.txt")
	txt, err = os.ReadFile(path)
	if err == nil {
		patterns = strings.Split(string(txt), "\n")
		for i := range patterns {
			patterns[i] = strings.TrimSpace(patterns[i])
		}
		acSize := ac.Setup(patterns)
		log.Info().Msgf("Created Title search trie with %d nodes.\n", acSize)
	}
	return ac, err
}
