package bhlnames

import (
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
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

	// NameRefs takes a name and optionally reference, and find matching
	// locations and references in BHL.
	NameRefs(data input.Input) (*namerefs.NameRefs, error)

	// NameRefsStream takes a stream of names/references and returns back
	// a stream of matched locations in BHL.
	NameRefsStream(chIn <-chan input.Input, chOut chan<- *namerefs.NameRefs)

	// NomenRefs takes a name and a nomenclatural reference and returns back
	// putative locations of the nomenclatural publication in BHL.
	NomenRefs(data input.Input) (*namerefs.NameRefs, error)

	// NomenRefsStream takes a stream of names/references and returns a stream
	// of putative locations of the nomenclatural publications in BHL.
	NomenRefsStream(chIn <-chan input.Input, chOut chan<- *namerefs.NameRefs)

	// GetVersion returns the version and build timestamp of bhlnames app.
	GetVersion() gnvers.Version

	// Config returns configuration data of bhlnames.
	Config() config.Config

	// ChangeConfig modifies config and returns back an instance of BHLnames
	// with updated Config.
	ChangeConfig(...config.Option) BHLnames

	// Close terminates connections to databases and key-value stores.
	Close() error
}
