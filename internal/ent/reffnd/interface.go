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

	// EmptyNameRefs returns empty non-nil result.
	EmptyNameRefs(inp input.Input) *bhl.RefsByName

	// RefByPageID returns a reference for a given pageID.
	RefByPageID(pageID int) (*bhl.Reference, error)

	// RefsByExtID returns references for a given external ID and data-source ID.
	// If allRefs is true, it returns all cached references for the external ID.
	// Otherwise it returns only the best match.
	RefsByExtID(
		extID string,
		data_source_id int,
	) (*bhl.RefsByName, error)

	// ItemStats returns metadata for a given itemID as well as the
	// taxonomic statistics for the item.
	ItemStats(itemID int) (*bhl.Item, error)

	// ItemsByTaxon returns a collection of BHL items that contain more than
	// 50% of the species of the profided taxon.
	ItemsByTaxon(taxon string) ([]*bhl.Item, error)

	// Close cleans up all the database, key-value store, files locks and blocks,
	// releasing resources for the next usage of the program.
	Close()
}
