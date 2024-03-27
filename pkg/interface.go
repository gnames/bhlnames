package bhlnames

import (
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/bhlnames/internal/ent/refbhl"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnlib/ent/gnvers"
	"github.com/gnames/gnparser"
)

// BHLnames provides methods for finding references for scientific names
// in the Biodiversity Heritage Library (BHL).
type BHLnames interface {
	// Parser returns an instance of the Global Names Parser. It is used
	// to parse scientific names.
	Parser() gnparser.GNparser

	// Initialize downloads data about BHL corpus and names found in the
	// corpus. It then imports the data into its storage.
	Initialize() error

	// InitializeCol downloads the Catalogue of Life (CoL) data and imports
	// the data about nomenclatural references for CoL names into its storage.
	InitializeCol() error

	// RefByPageID returns a reference for a given pageID.
	RefByPageID(pageID int) (*refbhl.Reference, error)

	// NameRefs takes a name and optionally reference, and find matching
	// locations and references in BHL.
	NameRefs(data input.Input) (*namerefs.NameRefs, error)

	// NameRefsStream takes a stream of names/references and returns back
	// a stream of matched locations in BHL.
	NameRefsStream(chIn <-chan input.Input, chOut chan<- *namerefs.NameRefs)

	// RefsByExternalID returns nomenclatural references corresponding to
	// a given data source and an external ID from it. If allRefs is true,
	// it returns all putative nomenclatural references for the external ID,
	// otherwise it returns only the best one.
	RefsByExternalID(
		dataSource, id string,
		allRefs bool,
	) ([]*refbhl.Reference, error)

	// GetVersion returns back the version of BHLnames
	// @Summary Get BHLnames version
	// @Description Retrieves the current version of the BHLnames application.
	// @ID get-version
	// @Produce json
	// @Success 200 {object} gnvers.Version "Successful response with version information"
	// @Router /version [get]
	GetVersion() gnvers.Version

	// Config returns configuration data of bhlnames.
	Config() config.Config

	// ChangeConfig modifies config and returns back an instance of BHLnames
	// with updated Config.
	ChangeConfig(...config.Option) BHLnames

	// Close terminates connections to databases and key-value stores.
	Close() error
}
