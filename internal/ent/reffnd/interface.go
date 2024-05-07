package reffnd

import (
	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/pkg/config"
)

// RefFinder interface contains methods to find BHL references according to
// input.
type RefFinder interface {
	// ReferencesByName takes input with name and a reference
	// and returns back back BHL references that match the input.
	ReferencesByName(
		inp input.Input,
		cfg config.Config,
	) (*bhl.RefsByName, error)

	// RefByPageID returns a reference for a given pageID.
	RefByPageID(pageID int) (*bhl.Reference, error)

	// RefsByExtID returns references for a given external ID and data-source ID.
	// If allRefs is true, it returns all cached references for the external ID.
	// Otherwise it returns only the best match.
	RefsByExtID(
		extID string,
		data_source_id int,
	) (*bhl.RefsByName, error)

	// Close cleans up all the database, key-value store, files locks and blocks,
	// releasing resources for the next usage of the program.
	Close()
}
