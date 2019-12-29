package bhlnames

type BHLnames struct {
	Rebuild      bool
	BHLdump      string
	BHLindexHost string
	InputDir     string
	DbHost       string
	DbUser       string
	DbPass       string
	DbName       string
	ProgressNum  int
}

// Option type for changing GNfinder settings.
type Option func(*BHLnames)

func OptBHLdump(d string) Option {
	return func(bhln *BHLnames) {
		bhln.BHLdump = d
	}
}

func OptBHLindexHost(bh string) Option {
	return func(bhln *BHLnames) {
		bhln.BHLindexHost = bh
	}
}

func OptInputDir(i string) Option {
	return func(bhln *BHLnames) {
		bhln.InputDir = i
	}
}

func OptDbHost(h string) Option {
	return func(bhln *BHLnames) {
		bhln.DbHost = h
	}
}

func OptDbUser(u string) Option {
	return func(bhln *BHLnames) {
		bhln.DbUser = u
	}
}

func OptDbPass(p string) Option {
	return func(bhln *BHLnames) {
		bhln.DbPass = p
	}
}

func OptDbName(n string) Option {
	return func(bhln *BHLnames) {
		bhln.DbName = n
	}
}

func OptProgressNum(n int) Option {
	return func(bhln *BHLnames) {
		bhln.ProgressNum = n
	}
}
func OptRebuild(r bool) Option {
	return func(bhln *BHLnames) {
		bhln.Rebuild = r
	}
}

func NewBHLnames(opts ...Option) BHLnames {
	bhln := BHLnames{}
	for _, opt := range opts {
		opt(&bhln)
	}
	return bhln
}
