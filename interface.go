package bhlnames

import (
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/builder"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/reffinder"
	"github.com/gnames/bhlnames/ent/title_matcher"
	"github.com/gnames/gnlib/ent/gnvers"
	"github.com/gnames/gnparser"
)

type BHLnames interface {
	builder.Builder
	reffinder.RefFinder
	title_matcher.TitleMatcher

	Parser() gnparser.GNparser

	Initialize() error

	NameRefs(data input.Input) (*namerefs.NameRefs, error)
	NameRefsStream(chIn <-chan input.Input, chOut chan<- *namerefs.NameRefs)

	NomenRefs(data input.Input) (*namerefs.NameRefs, error)
	NomenRefsStream(chIn <-chan input.Input, chOut chan<- *namerefs.NameRefs)

	GetVersion() gnvers.Version

	Config() config.Config
	ChangeConfig(...config.Option) BHLnames
	Close() error
}
