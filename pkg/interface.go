package bhlnames

import (
	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/builder"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnparser"
)

// BHLnames provides methods for finding references for scientific names
// in the Biodiversity Heritage Library (BHL).
type BHLnames interface {
	// Initialize downloads data about BHL corpus and names found in the
	// corpus. It then imports the data into its storage.
	Initialize(builder.Builder) error

	// NameRefs takes an input with a scientific name and, optionally, a
	// reference, and returns references found in BHL.
	NameRefs(input.Input) (*bhl.RefsByName, error)

	// NameRefsStream takes a channel with input that contains information
	// about scientific name and, optionally a reference, and returns a
	// channel with references found in BHL.
	NameRefsStream(chIn <-chan input.Input, chOut chan<- *bhl.RefsByName)

	// Config returns the configuration of the BHLnames instance.
	Config() config.Config

	// ParsePool return a buffered channel containing instances of GNparser.
	// It can be used as a pool of GNparser instances.
	ParserPool() chan gnparser.GNparser

	// Close terminates connections to databases and key-value stores.
	Close() error
}
