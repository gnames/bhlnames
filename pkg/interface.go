package bhlnames

import (
	"context"

	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/builder"
	"github.com/gnames/bhlnames/internal/ent/col"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnparser"
)

// BHLnames provides methods for finding references for scientific names
// in the Biodiversity Heritage Library (BHL).
type BHLnames interface {
	// Initialize downloads of essential BHL data (corpus metadata + names) and
	// prepares the internal storage for efficient querying.
	Initialize(builder.Builder) error

	// InitCoLNomenEvents fetches Catalogue of Life (CoL) data. It finds
	// nomenclatural events in BHL, cross-referencing them with names and
	// references from CoL.
	InitCoLNomenEvents(col.Nomen) error

	// NameRefs accepts a scientific name and optional reference. It returns a
	// collection of matching references found within the BHL corpus.
	NameRefs(input.Input) (*bhl.RefsByName, error)

	// NameRefsStream processes a stream of inputs (scientific names + optional
	// references). It returns a stream of corresponding reference collections
	// found in BHL. Designed for asynchronous processing and large-scale
	// requests.
	NameRefsStream(
		ctx context.Context,
		chIn <-chan input.Input,
		chOut chan<- *bhl.RefsByName,
	) error

	// Config returns the current configuration used by the BHLnames instance.
	Config() config.Config

	// ParserPool returns a channel for accessing reusable GNparser instances.
	// This allows efficient pooling and management of name-parser resources.
	ParserPool() chan gnparser.GNparser

	// Close releases all resources (database connections, etc.) used by BHLnames.
	Close()
}
