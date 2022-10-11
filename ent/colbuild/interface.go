// Package colbuild sets methods for loading names and references from
// the Catalogue of Life and then creating links from the references to
// Biodiversity Heritage Library locations.
package colbuild

import (
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
)

type ColBuild interface {
	// DataStatus determines if downloaded and extracted files do exist,
	// and if the tables exist and have data.
	DataStatus() (bool, bool, error)

	// ResetColData removes all downloaded and generated resources, leaving
	// empty databases and no files.
	ResetColData()

	// ImportColData downloads CoL Darwin Core Archive dump, then imports names
	// and references from into storage.
	ImportColData() error

	// LinkColToBhl discovers putative links from CoL references to BHL pages
	// and stores the results.
	LinkColToBhl(func(<-chan input.Input, chan<- *namerefs.NameRefs)) error
}
