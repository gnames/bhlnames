package data

import (
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/domain/entity"
)

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

type Builder interface {
}
