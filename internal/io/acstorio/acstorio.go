package acstorio

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gnames/bhlnames/internal/ent/acstor"
	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/gnames/bhlnames/internal/io/dbio"
	"github.com/gnames/bhlnames/internal/io/dictio"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/bhlnames/pkg/ent/abbr"
	"github.com/jackc/pgx/v5/pgxpool"
)

type acstorio struct {
	cfg        config.Config
	titles     map[int]*model.Title
	shortWords map[string]struct{}
	abbrMap    map[string][]int
	db         *pgxpool.Pool
}

// New creates a new instance of AhoCorasickStore.
func New(
	cfg config.Config,
	db *pgxpool.Pool,
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
		db:         db,
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

// Get returns a list of title IDs that contain the key.
func (a *acstorio) Get(abbr string) ([]int, error) {
	var res []int
	ctx := context.Background()
	q := `SELECT title_id FROM abbr_titles WHERE abbr = $1`
	rows, err := a.db.Query(ctx, q, abbr)
	if err != nil {
		slog.Error("Cannot get title IDs for abbreviation",
			"abbr", abbr,
			"error", err,
		)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			slog.Error("Cannot scan title ID for abbreviation",
				"abbr", abbr,
				"error", err,
			)
			return nil, err
		}
		res = append(res, id)
	}
	return res, nil
}

// save stores a map of abbreviasions as keys and title IDs as values
// in a key-value store. It also saves the list of all abbreviations
// in a file. This file is used later to generate an Aho-Corasick trie.
func (a *acstorio) save(abbrMap map[string][]int) error {
	err := a.saveAbbrTitle(abbrMap)
	if err != nil {
		slog.Error("Cannot save abbreviations to DB", "error", err)
		return err
	}

	err = a.saveAbbr(abbrMap)
	if err != nil {
		slog.Error("Cannot save abbreviations to file", "error", err)
		return err
	}
	return nil
}

func (a *acstorio) saveAbbrTitle(abbrMap map[string][]int) error {
	slog.Info("Saving maps of abbreviations to titles.")
	columns := []string{"abbr", "title_id"}
	var rows [][]interface{}
	for k, v := range abbrMap {
		for i := range v {
			rows = append(rows, []interface{}{k, v[i]})
		}
	}
	_, err := dbio.InsertRows(a.db, "abbr_titles", columns, rows)
	if err != nil {
		slog.Error("Cannot save abbreviations to DB", "error", err)
		return err
	}
	return nil
}

func (a *acstorio) saveAbbr(abbrMap map[string][]int) error {
	slog.Info("Saving stand-alone abbreviations.")
	columns := []string{"abbr"}
	rows := make([][]interface{}, len(abbrMap))
	var count int
	for k := range abbrMap {
		rows[count] = []interface{}{k}
		count++
	}
	_, err := dbio.InsertRows(a.db, "abbrs", columns, rows)
	if err != nil {
		slog.Error("Cannot save abbreviations to DB", "error", err)
		return err
	}
	return nil
}
