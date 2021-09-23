package bhlnames

import (
	"github.com/gnames/bhlnames/ent/builder"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/reffinder"
	"github.com/gnames/gnfmt"
)

type BHLnames interface {
	Initialize(b builder.Builder) error
	NameRefs(rf reffinder.RefFinder, data input.Input) (*namerefs.NameRefs, error)
	NameRefsStream(rf reffinder.RefFinder, chIn <-chan input.Input, chOut chan<- *namerefs.NameRefs)
	NomenRefs(rf reffinder.RefFinder, data input.Input) (*namerefs.NameRefs, error)
	NomenRefsStream(
		rf reffinder.RefFinder,
		chIn <-chan input.Input,
		chOut chan<- *namerefs.NameRefs,
	)
	Format() gnfmt.Format
}
