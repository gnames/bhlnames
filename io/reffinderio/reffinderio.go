package reffinderio

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/aho_corasick"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/refbhl"
	"github.com/gnames/bhlnames/ent/reffinder"
	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/gnparser"
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog/log"
)

// reffinderio is an implementation of Librarian interface. It uses
// pre-populated PostgreSQL database and Badger key-value store to find
// BHL references where a scientific name-string appears.
type reffinderio struct {
	// withSynonyms is true if searches should be augmented with
	// synonyms of a name.
	withSynonyms bool

	// sortDesc is a flag that tells to sort data in descendent order.
	sortDesc bool

	// withShortenedOutput is a flag to return output with a summary only,
	// skipping details about found references.
	withShortenedOutput bool

	// KV contains a key-value Badger store where data about known
	// publications is kept.
	KV *badger.DB

	// DB is a PostgreSQL connection for plain SQL-queries.
	DB *sql.DB

	// GormDB is a PostgreSQL connection for ORM-queries.
	GormDB *gorm.DB

	// AC is AhoCorasick object for matching references to BHL titles.
	AC aho_corasick.AhoCorasick
}

func New(cfg config.Config) reffinder.RefFinder {
	log.Info().Msgf("Connecting to PostgreSQL database %s at %s", cfg.DbDatabase, cfg.DbHost)
	res := &reffinderio{
		KV:     db.InitKeyVal(cfg.PartDir),
		DB:     db.NewDB(cfg),
		GormDB: db.NewDbGorm(cfg),
	}
	return res
}

func (rf reffinderio) ReferencesBHL(
	inp input.Input,
	cfg config.Config) (*namerefs.NameRefs, error) {
	var err error
	rf.withSynonyms = cfg.WithSynonyms
	rf.sortDesc = cfg.SortDesc
	rf.withShortenedOutput = cfg.WithShortenedOutput

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

	var rows []*row
	if rf.withSynonyms {
		rows = rf.taxonOccurrences(res)
	} else {
		rows = rf.nameOnlyOccurrences(res)
	}

	res.ImagesURL = imagesUrl(res.CurrentCanonical)
	rf.updateOutput(res, rows)
	return res, nil
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

func (rf reffinderio) emptyNameRefs(data input.Input) *namerefs.NameRefs {
	res := &namerefs.NameRefs{
		Input:        data,
		References:   make([]*refbhl.ReferenceBHL, 0),
		WithSynonyms: rf.withSynonyms,
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
