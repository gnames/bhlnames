package reffinderio

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/aho_corasick"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/abbr"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/refbhl"
	"github.com/gnames/bhlnames/ent/reffinder"
	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnparser"
	"github.com/jinzhu/gorm"
)

// reffinderio is an implementation of Librarian interface. It uses
// pre-populated PostgreSQL database and Badger key-value store to find
// BHL references where a scientific name-string appears.
type reffinderio struct {

	// Config contains general configuration of BHLnames. Some of the config
	// settings modify behavior of algorithms to find BHL references.
	config.Config

	// KV contains a key-value Badger store where data about known
	// publications is kept.
	KV *badger.DB

	// TitleKV contains a key-value Badger store where data about titleIDs and
	// title abbreviations is kept.
	TitleKV *badger.DB

	// DB is a PostgreSQL connection for plain SQL-queries.
	DB *sql.DB

	// GormDB is a PostgreSQL connection for ORM-queries.
	GormDB *gorm.DB

	// AC is AhoCorasick object for matching references to BHL titles.
	AC aho_corasick.AhoCorasick
}

func New(cfg config.Config) (reffinder.RefFinder, error) {
	log.Printf("Connecting to PostgreSQL database %s at %s", cfg.DbName, cfg.DbHost)
	res := reffinderio{
		Config:  cfg,
		KV:      db.InitKeyVal(cfg.PartDir),
		TitleKV: db.InitKeyVal(cfg.AhoCorKeyValDir),
		DB:      db.NewDB(cfg),
		GormDB:  db.NewDbGorm(cfg),
	}
	ac, err := res.getAhoCorasick()
	if err != nil {
		return res, err
	}
	res.AC = ac
	return res, nil
}

func (rf reffinderio) TitlesBHL(refString string) (map[int][]string, error) {
	refAbbr := abbr.Abbr(refString)
	matches := rf.AC.Search(refAbbr)

	abbrs := make([]string, len(matches))
	for i := range matches {
		abbrs[i] = matches[i].Pattern
	}
	return rf.abbrsToTitleIDs(abbrs)
}

func (rf reffinderio) ReferencesBHL(data input.Input) (*namerefs.NameRefs, error) {
	var err error
	// gets empty *namerefs.NameRefs with current_canonical
	res := rf.emptyNameRefs(data)

	res.Canonical, err = fullCanonical(data.NameString)
	if err != nil {
		return res, err
	}
	res.CurrentCanonical, err = rf.currentCanonical(res.Canonical)
	if err != nil {
		return res, err
	}

	var rows []*row
	if !rf.WithSynonyms {
		rows = rf.nameOnlyOccurrences(res)
	} else {
		rows = rf.taxonOccurrences(res)
	}

	res.ImagesURL = imagesUrl(res.CurrentCanonical)

	rf.updateOutput(res, rows)
	return res, nil
}

func (rf reffinderio) Close() error {
	err1 := rf.DB.Close()
	err2 := rf.GormDB.Close()
	err3 := rf.KV.Close()
	err4 := rf.TitleKV.Close()

	for _, err := range []error{err1, err2, err3, err4} {
		if err != nil {
			return err
		}
	}
	return nil
}

func (rf reffinderio) getAhoCorasick() (aho_corasick.AhoCorasick, error) {
	var err error
	var txt []byte
	var patterns []string
	ac := aho_corasick.New()

	path := filepath.Join(rf.AhoCorasickDir, "patterns.txt")
	txt, err = os.ReadFile(path)
	if err == nil {
		patterns = strings.Split(string(txt), "\n")
		ac.Setup(patterns)
	}
	return ac, err
}

func (rf reffinderio) abbrsToTitleIDs(abbrs []string) (map[int][]string, error) {
	var ids []string
	res := make(map[int][]string)
	enc := gnfmt.GNgob{}

	vals, err := db.GetValues(rf.TitleKV, abbrs)
	if err != nil {
		return res, err
	}

	for k, v := range vals {
		err = enc.Decode(v, ids)
		if err != nil {
			return res, err
		}
		for i := range ids {
			titleID, err := strconv.Atoi(ids[i])
			if err != nil {
				return res, err
			}
			res[titleID] = append(res[titleID], k)
		}
	}
	return res, nil
}

func (rf reffinderio) emptyNameRefs(data input.Input) *namerefs.NameRefs {
	res := &namerefs.NameRefs{
		Input:        data,
		References:   make([]*refbhl.ReferenceBHL, 0),
		WithSynonyms: rf.WithSynonyms,
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
