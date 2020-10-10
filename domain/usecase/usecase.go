package usecase

import (
	linkent "github.com/gdower/bhlinker/domain/entity"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/domain/entity"
)

// Librarian interface is a lower-level one, that tells how to find
// BHL references for name_strings.
type Librarian interface {
	// ReferencesBHL takes a name-string and returns back back BHL references
	// where this name-string was detected.  If a name-string occurrence has a
	// nomenclatural annotaion (like 'sp. nov.') attached to it somewhere in the
	// reference, the position of that occurrence is returned. If there is no
	// such annotation, the first detection of the name-string in the reference
	// is returned.  When a reference is an `item` (a journal volume usually)
	// with `parts` (a publication/article usually), we return one occurrence for
	// every `part`, but also first occurrence of a name-string in the `item`, if
	// it exists outside of all `parts`.
	ReferencesBHL(name_string string, opts ...config.Option) (*entity.NameRefs, error)

	// Close cleans up all the database, key-value store, files locks and blocks,
	// releasing resources for the next usage of the program.
	Close() error
}

// Builder uses remote resources to recreate all necessary data from scratch
// locally.
type Builder interface {
	// Reset data removes all downloaded and generated resources, leaving
	// empty databases and directories.
	ResetData()
	// Import downloads remote datasets to the local file system and generates
	// data needed for bhlnames functionality.
	ImportData() error
}

// APIProvider is implemented by API like REST.
type APIProvider interface {
	// Port provides the port number of a service
	Port() int
	// NameRefs takes a slice of scientific name-strings and returns back
	// BHL refrences that correspond to these name-strings. It does not
	// try to match currnetly accepted name or synonyms of a taxa the name-string
	// is pointing to.
	NameRefs(nameStrings []string) []*entity.NameRefs
	// TaxonRefs takes a slice of scientific name-strings and returns back
	// BHL references that mention either currently accepted name or any
	// synonym that belongs to the corrsponsing taxon.
	TaxonRefs(nameStrings []string) []*entity.NameRefs
	// NomenRrefs take a slice of a scientific name/original publication pair
	// and tries to find the publication in BHL.
	NomenRefs(inputs []linkent.Input) []linkent.Output
}
