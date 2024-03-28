package acstorio

import (
	"cmp"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/internal/ent/acstor"
	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/gnames/bhlnames/internal/io/dbio"
	"github.com/gnames/bhlnames/internal/io/dictio"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/bhlnames/pkg/ent/abbr"
	"github.com/gnames/gnfmt"
)

type acstorio struct {
	cfg        config.Config
	titles     map[int]*model.Title
	shortWords map[string]struct{}
	abbrMap    map[string][]int
	db         *badger.DB
}

// New creates a new instance of AhoCorasickStore.
func New(
	cfg config.Config,
	titleMap map[int]*model.Title,
) (acstor.AhoCorasickStore, error) {
	d := dictio.New()
	shortWords, err := d.ShortWords()
	if err != nil {
		slog.Error("Cannot generate short words for AhoCoraickStore", "error", err)
		return nil, err
	}
	res := acstorio{
		cfg:        cfg,
		titles:     titleMap,
		shortWords: shortWords,
	}
	return &res, nil
}

// Setup prepares the AhoCorasickStore for use.
func (a *acstorio) Setup() error {
	slog.Info("Setting up AhoCorasickStore.")
	// abbrMap maps abbreviation strings to the list of
	// title IDs that contain the abbreviation
	abbrMap := make(map[string][]int)

	if a.titles == nil {
		err := errors.New("titles data is nil")
		slog.Error("Cannot setup AhoCorasickStore.", "error", err)
		return err
	}

	for k, v := range a.titles {
		abbrs := abbr.Patterns(v.Name, a.shortWords)
		for i := range abbrs {
			if len(abbrs[i]) > 2 {
				abbrMap[abbrs[i]] = append(abbrMap[abbrs[i]], k)
			}
		}
	}
	return a.save(abbrMap)
}

// Open opens the AhoCorasickStore for use.
func (a *acstorio) Open() error {
	slog.Info("Opening AhoCorasickStore.")
	db, err := dbio.InitKeyVal(a.cfg.AhoCorKeyValDir, true)
	if err != nil {
		slog.Error("Cannot open key-value for AhoCorasickStore", "error", err)
		return err
	}
	a.db = db

	return nil
}

// Get returns a list of title IDs that contain the key.
func (a *acstorio) Get(abbr string) ([]int, error) {
	vals, err := dbio.GetValues(a.db, []string{abbr})
	if err != nil {
		slog.Error("Cannot get values from AhoCorasickStore", "error", err)
		return nil, err
	}
	if titleIDsGob, ok := vals[abbr]; ok {
		var titleIDs []int
		enc := gnfmt.GNgob{}
		err = enc.Decode(titleIDsGob, &titleIDs)
		if err != nil {
			slog.Error("Cannot decode values from AhoCorasickStore", "error", err)
			return nil, err
		}
		return titleIDs, nil
	}
	return nil, nil
}

// save stores a map of abbreviasions as keys and title IDs as values
// in a key-value store. It also saves the list of all abbreviations
// in a file. This file is used later to generate an Aho-Corasick trie.
func (a *acstorio) save(abbrMap map[string][]int) error {
	var err error
	kv, err := dbio.InitKeyVal(a.cfg.AhoCorKeyValDir, false)
	if err != nil {
		slog.Error("Cannot open key-value for AhoCorasickStore", "error", err)
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
				slog.Error(
					"Cannot commit intermediate transaction for AhoCorasickStore",
					"error", err,
				)
				return err
			}

			kvTxn = kv.NewTransaction(true)
			err = kvTxn.Set([]byte(k), bs)
			if err != nil {
				slog.Error("Cannot set key-value for AhoCorasickStore", "error", err)
				return err
			}
		}
	}
	err = kvTxn.Commit()
	if err != nil {
		slog.Error(
			"Cannot commit transaction for AhoCorasickStore",
			"error", err,
		)
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

	// create a file with all abbreviations for Aho-Corasick trie generation
	tmp := strings.Join(abbrs, "\n")
	path := filepath.Join(a.cfg.AhoCorasickDir, "patterns.txt")
	err = os.WriteFile(path, []byte(tmp), 0644)
	if err != nil {
		slog.Error("Cannot write patterns.txt for AhoCorasickStore", "error", err)
		return err
	}

	return nil
}
