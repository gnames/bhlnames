package col

import (
	"context"

	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/input"
)

// Nomen provides methods for working with Catalogue of Life (CoL) data and
// extracting nomenclatural events from the Biodiversity Heritage Library (BHL).
// It supports efficient retrieval of these events using scientific names and
// their references.
type Nomen interface {
	// CheckCoLData verifies if the CoL archive file exists and if its contents
	// have been previously extracted. Returns flags indicating their status.
	// It also checks if there are CoL-related records in the database.
	CheckCoLData() (bool, bool, error)

	// ResetCoLData removes all CoL-related data (downloaded files, generated
	// resources). This restores the system to a clean state with no CoL data.
	ResetCoLData() error

	// ImportCoLData downloads the CoL Darwin Core Archive and imports relevant
	// taxonomic names and references into the internal storage.
	ImportCoLData() error

	// NomenEvents locates putative nomenclatural events in BHL associated with
	// names from the Catalogue of Life (CoL). This method leverages a provided
	// function to process input names and return BHL references.
	NomenEvents(
		func(context.Context, <-chan input.Input, chan<- *bhl.RefsByName) error,
	) error

	// Close releases all resources (e.g., database connections) used by the
	// Nomen instance.
	Close()
}
