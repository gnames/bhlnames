package reffndio

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/gnames/aho_corasick"
	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/reffnd"
	"github.com/gnames/bhlnames/internal/io/dbio"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnparser"
	"github.com/jackc/pgx/v5/pgxpool"
)

type reffndio struct {
	// db is a PostgreSQL connection for plain SQL-queries.
	db *pgxpool.Pool

	// AC is AhoCorasick object for matching references to BHL titles.
	AC aho_corasick.AhoCorasick
}

func New(cfg config.Config) (reffnd.RefFinder, error) {
	slog.Info("Connecting to PostgreSQL database", "database", cfg.DbDatabase, "host", cfg.DbHost)

	dbConn, err := dbio.NewDB(cfg)
	if err != nil {
		return nil, err
	}

	res := &reffndio{
		db: dbConn,
	}
	return res, nil
}

func (rf reffndio) ReferencesByName(
	inp input.Input,
	cfg config.Config) (*bhl.RefsByName, error) {
	var err error

	// gets empty *namerefs.NameRefs with current_canonical
	res := rf.emptyNameRefs(inp)

	res.Canonical, err = fullCanonical(inp.NameString)
	if err != nil {
		return res, err
	}
	res.CurrentCanonical, err = rf.currentCanonical(res.Canonical)
	if err != nil {
		return res, err
	}

	var refRecs []*refRec
	if inp.WithTaxon {
		refRecs, err = rf.taxonOccurrences(res)
		if err != nil {
			return nil, err
		}
	} else {
		refRecs, err = rf.nameOnlyOccurrences(res)
		if err != nil {
			return nil, err
		}
	}
	res.ImagesURL = imagesUrl(res.CurrentCanonical)
	rf.deduplicateResults(inp, res, refRecs)
	return res, nil
}

func (r *reffndio) RefByPageID(pageID int) (*bhl.Reference, error) {
	return nil, nil
}

func (r *reffndio) Close() {
	r.db.Close()
}

func (rf *reffndio) emptyNameRefs(inp input.Input) *bhl.RefsByName {
	meta := bhl.Meta{
		Input: inp,
	}
	res := &bhl.RefsByName{
		Meta:       meta,
		References: make([]*bhl.ReferenceName, 0),
	}
	return res
}

func fullCanonical(name_string string) (string, error) {
	cfg := gnparser.NewConfig()
	gnp := gnparser.New(cfg)
	ps := gnp.ParseName(name_string)
	if !ps.Parsed {
		return "", fmt.Errorf("could not parse name_string '%s'", name_string)
	}
	can := ps.Canonical.Simple
	return can, nil
}

func imagesUrl(name string) string {
	q := url.PathEscape(name)
	url := "https://www.google.com/search?tbm=isch&q=%s"
	return fmt.Sprintf(url, q)
}
