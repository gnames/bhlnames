package bhlnames

type BHLnames struct {
	DbHost   string
	User     string
	Password string
}

// Option type for changing GNfinder settings.
type Option func(*BHLnames)

func OptDbHost(h string) Option {
	return func(bhln *BHLnames) {
		bhln.DbHost = h
	}
}

func NewBHLnames(opts ...Option) BHLnames {
	bhln := BHLnames{}
	for _, opt := range opts {
		opt(&bhln)
	}
	return bhln
}
