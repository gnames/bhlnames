package bhlnames

import (
	"github.com/gnames/bhlnames/bhl"
	"github.com/gnames/bhlnames/db"
)

type BHLnames struct {
	db.DbOpts
	bhl.MetaData
	Format     string
	JobsNum    int
	SortDesc   bool
	Short      bool
	NoSynonyms bool
}

// Option type for changing GNfinder settings.
type Option func(*BHLnames)

func OptDumpURL(d string) Option {
	return func(bhln *BHLnames) {
		bhln.MetaData.DumpURL = d
	}
}

func OptBHLindexHost(bh string) Option {
	return func(bhln *BHLnames) {
		bhln.BHLindexHost = bh
	}
}

func OptInputDir(i string) Option {
	return func(bhln *BHLnames) {
		bhln.MetaData.InputDir = i
	}
}

func OptDbHost(h string) Option {
	return func(bhln *BHLnames) {
		bhln.DbOpts.Host = h
	}
}

func OptDbUser(u string) Option {
	return func(bhln *BHLnames) {
		bhln.DbOpts.User = u
	}
}

func OptDbPass(p string) Option {
	return func(bhln *BHLnames) {
		bhln.DbOpts.Pass = p
	}
}

func OptDbName(n string) Option {
	return func(bhln *BHLnames) {
		bhln.DbOpts.Name = n
	}
}

func OptFormat(f string) Option {
	return func(bhln *BHLnames) {
		f = checkFormat(f)
		bhln.Format = f
	}
}

func OptRebuild(r bool) Option {
	return func(bhln *BHLnames) {
		bhln.Rebuild = r
	}
}

func OptJobsNum(j int) Option {
	return func(bhln *BHLnames) {
		bhln.JobsNum = j
	}
}

func OptSortDesc(d bool) Option {
	return func(bhln *BHLnames) {
		bhln.SortDesc = d
	}
}

func OptShort(s bool) Option {
	return func(bhln *BHLnames) {
		bhln.Short = s
	}
}

func OptNoSynonyms(n bool) Option {
	return func(bhln *BHLnames) {
		bhln.NoSynonyms = n
	}
}

func NewBHLnames(opts ...Option) BHLnames {
	bhln := BHLnames{}
	for _, opt := range opts {
		opt(&bhln)
	}
	bhln.MetaData.Configure(bhln.DbOpts)
	return bhln
}

func checkFormat(f string) string {
	if f != "compact" && f != "pretty" {
		return "compact"
	}
	return f
}
