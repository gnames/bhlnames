package ttlmchio

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gnames/aho_corasick"
	"github.com/gnames/bhlnames/internal/ent/ttlmch"
	"github.com/gnames/bhlnames/internal/io/dbio"
	"github.com/gnames/bhlnames/internal/io/dictio"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ttlmchio struct {
	// AC is AhoCorasick object for matching references to BHL titles.
	aho_corasick.AhoCorasick

	shortWords map[string]struct{}

	// db is a connection to the database.
	db *pgxpool.Pool
}

func New(cfg config.Config) (ttlmch.TitleMatcher, error) {
	db, err := dbio.NewDB(cfg)
	if err != nil {
		slog.Error(
			"Cannot create database connection for TitleMatcher",
			"error", err,
		)
		return nil, err
	}

	d := dictio.New()
	shortWords, err := d.ShortWords()
	if err != nil {
		slog.Error("Cannot get short words", "error", err)
		return nil, err
	}

	res := ttlmchio{
		db:         db,
		shortWords: shortWords,
	}

	ac, err := res.getAhoCorasick()
	if err != nil {
		err = fmt.Errorf("titlemio.New: %w", err)
		slog.Error("Cannot create AhoCorasick", "error", err)
		return nil, err
	}
	res.AhoCorasick = ac

	return &res, nil
}

func (tm *ttlmchio) Close() {
	tm.db.Close()
}

func (tm ttlmchio) getAhoCorasick() (aho_corasick.AhoCorasick, error) {
	var err error
	ac := aho_corasick.New()

	q := `SELECT abbr from abbrs`
	rows, err := tm.db.Query(context.Background(), q)
	if err != nil {
		slog.Error("Cannot get abbreviations from the database", "error", err)
		return ac, err
	}
	defer rows.Close()

	patterns := make([]string, 0)
	for rows.Next() {
		var abbr string
		err = rows.Scan(&abbr)
		if err != nil {
			slog.Error("Cannot scan abbreviation", "error", err)
			return ac, err
		}
		patterns = append(patterns, abbr)
	}
	acSize := ac.Setup(patterns)
	slog.Info("Created Title search trie", "trie_size", acSize)
	return ac, err
}
