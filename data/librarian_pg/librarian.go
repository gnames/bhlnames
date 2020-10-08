package librarian_pg

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/db"
	"github.com/gnames/bhlnames/domain/entity"
	"github.com/jinzhu/gorm"
	"gitlab.com/gogna/gnparser"
)

// LibrarianPG is an implementation of Librarian interface. It uses
// pre-populated PostgreSQL database and Badger key-value store to find
// BHL references where a scientific name-string appears.
type LibrarianPG struct {

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

func NewLibrarianPG(cfg config.Config) LibrarianPG {
	res := LibrarianPG{
		Config: cfg,
		KV:     db.InitKeyVal(cfg.PartDir),
		DB:     db.NewDb(cfg.DB),
		GormDB: db.NewDbGorm(cfg.DB),
	}
	return res
}

func (l LibrarianPG) Close() error {
	err1 := l.DB.Close()
	err2 := l.GormDB.Close()
	err3 := l.KV.Close()

	for _, err := range []error{err1, err2, err3} {
		if err != nil {
			return err
		}
	}
	return nil
}

func (l LibrarianPG) ReferencesBHL(name_string string, opts ...config.Option) (*entity.NameRefs, error) {
	for _, opt := range opts {
		opt(&l.Config)
	}
	var err error
	// gets empty *entity.NameRefs with current_canonical
	res := l.emptyNameRefs(name_string)

	res.Canonical, err = fullCanonical(name_string)
	if err != nil {
		return res, err
	}
	res.CurrentCanonical, err = l.currentCanonical(res.Canonical)
	if err != nil {
		return res, err
	}

	var rows []*row
	if l.NoSynonyms {
		rows = l.nameOnlyOccurrences(res)
	} else {
		rows = l.taxonOccurrences(res)
	}

	res.ImagesUrl = imagesUrl(res.CurrentCanonical)

	l.updateOutput(res, rows)
	return res, nil
}

func (l LibrarianPG) emptyNameRefs(name_string string) *entity.NameRefs {
	res := &entity.NameRefs{
		NameString:       name_string,
		Canonical:        "",
		CurrentCanonical: "",
		ImagesUrl:        "",
		References:       make([]*entity.Reference, 0),
		Params:           l.Config.RefParams,
	}
	return res
}

func fullCanonical(name_string string) (string, error) {
	gnp := gnparser.NewGNparser()
	ps := gnp.ParseToObject(name_string)
	if !ps.Parsed {
		return "", fmt.Errorf("could not parse name_string '%s'", name_string)
	}
	can := ps.Canonical.GetFull()
	return can, nil
}

func imagesUrl(name string) string {
	q := url.PathEscape(name)
	url := "https://www.google.com/search?tbm=isch&q=%s"
	return fmt.Sprintf(url, q)
}
