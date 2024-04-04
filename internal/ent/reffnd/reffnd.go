package reffnd

import (
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/pkg/config"
)

// RefFinder interface contains methods to find BHL references according to
// input.
type RefFinder interface {
	// ReferencesBHL takes input with name and a reference
	// and returns back back BHL references that match the input.
	ReferencesBHL(
		inp input.Input,
		cfg config.Config,
	) (*namerefs.NameRefs, error)

	// RefByPageID returns a reference for a given pageID.
	RefByPageID(pageID int) (*bhl.Reference, error)

	// Close cleans up all the database, key-value store, files locks and blocks,
	// releasing resources for the next usage of the program.
	Close() error
}
