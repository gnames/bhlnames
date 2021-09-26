package input

import (
	"strings"

	"github.com/gnames/gnparser"
)

type Input struct {
	ID        string `json:"id"`
	Name      `json:"name"`
	Reference `json:"reference"`
}

type Name struct {
	NameString string `json:"nameString,omitempty"`
	NameYear   string `json:"year,omitempty"`
	Canonical  string `json:"canonical,omitempty"`
	Authorship string `json:"authorship,omitempty"`
}

type Reference struct {
	RefString string `json:"refString,omitempty"`
	RefYear   string `json:"year,omitempty"`
	Authors   string `json:"authors,omitempty"`
	Journal   string `json:"journal,omitempty"`
	Volume    string `json:"volume,omitempty"`
	Pages     string `json:"pages,omitempty"`
}

type Option func(*Input)

func OptNameString(s string) Option {
	return func(i *Input) {
		i.NameString = s
	}
}

func OptNameYear(s string) Option {
	return func(i *Input) {
		i.NameYear = s
	}
}

func New(gnp gnparser.GNparser, opts ...Option) Input {
	res := Input{}
	for i := range opts {
		opts[i](&res)
	}

	if res.NameString != "" && res.Canonical == "" {
		parsed := gnp.ParseName(res.NameString)
		res.Canonical = parsed.Canonical.Simple
		res.Authors = strings.Join(parsed.Authorship.Authors, " ")
		if res.NameYear == "" {
			res.NameYear = parsed.Authorship.Year
		}
	}
	return res
}
