package reffndio

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/gnames/aho_corasick"
	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/reffnd"
	"github.com/gnames/bhlnames/internal/io/dbio"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnparser"
	"github.com/jackc/pgx/v5/pgxpool"
)

type reffndio struct {
	// db is a PostgreSQL connection for plain SQL-queries.
	db *pgxpool.Pool

	// ac is AhoCorasick object for matching references to BHL titles.
	ac aho_corasick.AhoCorasick

	// enc is an encoder/decoder for serializing/deserializing data.
	enc gnfmt.Encoder

	// ctx is a placeholder for database queries, for now it is "empty".
	ctx context.Context
}

func New(cfg config.Config) (reffnd.RefFinder, error) {
	slog.Info("Connecting to PostgreSQL database", "database", cfg.DbDatabase, "host", cfg.DbHost)

	dbConn, err := dbio.NewDB(cfg)
	if err != nil {
		return nil, err
	}

	res := &reffndio{
		db:  dbConn,
		enc: gnfmt.GNgob{},
		ctx: context.Background(),
	}
	return res, nil
}

func (rf reffndio) ReferencesByName(
	inp input.Input,
	cfg config.Config) (*bhl.RefsByName, error) {
	var err error

	// gets empty *namerefs.NameRefs with current_canonical
	res := rf.EmptyNameRefs(inp)

	res.Canonical, _ = simpleCanonical(inp.NameString)
	res.CurrentCanonical, err = rf.currentCanonical(res.Canonical)
	if err != nil {
		slog.Error(
			"Could not get current canonical",
			"canonical", res.Canonical,
			"error", err,
		)
	}

	var refRecs []*refRec

	if inp.Reference == nil && inp.WithNomenEvent {
		res, err = rf.colNomen(inp)
		if err != nil {
			slog.Error("Cannot get nomenclatural reference from CoL", "error", err)
			return nil, err
		}
		return res, nil
	}

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
	rf.deduplicateResults(inp, res, refRecs)
	return res, nil
}

func (rf *reffndio) RefByPageID(pageID int) (*bhl.Reference, error) {
	var ref *bhl.Reference
	ref, err := rf.refByPageID(pageID)
	if err != nil {
		return ref, err
	}
	return ref, nil
}

func (rf *reffndio) RefsByExtID(
	extID string,
	dataSourceID int,
) (*bhl.RefsByName, error) {
	_ = dataSourceID // not used yet, here for future use

	bs, err := rf.refsByExtID(extID)
	if err != nil {
		return nil, err
	}
	if bs == nil {
		return nil, nil
	}

	var res bhl.RefsByName
	err = rf.enc.Decode(bs, &res)
	if err != nil {
		slog.Error("Cannot decode refs by external ID", "error", err)
		return nil, err
	}
	return &res, nil
}

func (rf *reffndio) ItemStats(itemID int) (*bhl.Item, error) {
	res, err := rf.itemStats(itemID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ItemsByTaxon returns a collection of BHL items that contain more than
// 50% of the species of the profided taxon.
func (rf *reffndio) ItemsByTaxon(taxon string) ([]*bhl.Item, error) {
	items, err := rf.itemsByTaxon(taxon)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (rf *reffndio) Close() {
	rf.db.Close()
}

func (rf *reffndio) EmptyNameRefs(inp input.Input) *bhl.RefsByName {
	meta := bhl.Meta{
		Input: inp,
	}
	res := &bhl.RefsByName{
		Meta:       meta,
		References: make([]*bhl.ReferenceName, 0),
	}
	return res
}

func simpleCanonical(name_string string) (string, error) {
	cfg := gnparser.NewConfig()
	gnp := gnparser.New(cfg)
	ps := gnp.ParseName(name_string)
	if !ps.Parsed {
		err := errors.New("cannot parse name_string")
		slog.Error(
			"Name parsing error",
			"name_string", name_string,
			"error", err,
		)
		return "", err
	}
	can := ps.Canonical.Simple
	return can, nil
}

func imagesUrl(name string) string {
	q := url.PathEscape(name)
	url := "https://www.google.com/search?tbm=isch&q=%s"
	return fmt.Sprintf(url, q)
}
