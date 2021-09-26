package reffinderio

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/refbhl"
	"github.com/gnames/bhlnames/io/db"
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

	// DB is a PostgreSQL connection for plain SQL-queries.
	DB *sql.DB

	// GormDB is a PostgreSQL connection for ORM-queries.
	GormDB *gorm.DB
}

func New(cfg config.Config) reffinderio {
	log.Printf("Connecting to PostgreSQL database %s at %s", cfg.DbName, cfg.DbHost)
	res := reffinderio{
		Config: cfg,
		KV:     db.InitKeyVal(cfg.PartDir),
		DB:     db.NewDB(cfg),
		GormDB: db.NewDbGorm(cfg),
	}
	return res
}

func (rf reffinderio) Close() error {
	err1 := rf.DB.Close()
	err2 := rf.GormDB.Close()
	err3 := rf.KV.Close()

	for _, err := range []error{err1, err2, err3} {
		if err != nil {
			return err
		}
	}
	return nil
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
