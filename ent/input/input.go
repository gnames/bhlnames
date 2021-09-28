package input

import (
	"strings"

	"github.com/gnames/gnparser"
	"github.com/google/uuid"
)

type Input struct {
	ID        string `json:"id"`
	Name      `json:"name"`
	Reference `json:"reference"`
}

type Name struct {
	NameString  string `json:"nameString,omitempty"`
	NameYear    string `json:"year,omitempty"`
	Canonical   string `json:"canonical,omitempty"`
	NameAuthors string `json:"authors,omitempty"`
}

type Reference struct {
	RefString  string `json:"refString,omitempty"`
	RefYear    string `json:"year,omitempty"`
	RefAuthors string `json:"authors,omitempty"`
	Journal    string `json:"journal,omitempty"`
	Volume     string `json:"volume,omitempty"`
	PageStart  int    `json:"pageStart,omitempty"`
	PageEnd    int    `json:"pageEnd"`
}

type Option func(*Input)

func OptID(s string) Option {
	return func(i *Input) {
		i.ID = s
	}
}

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

func OptRefString(s string) Option {
	return func(i *Input) {
		i.RefString = s
	}
}

func New(gnp gnparser.GNparser, opts ...Option) Input {
	res := Input{}
	for i := range opts {
		opts[i](&res)
	}

	if res.ID == "" {
		res.ID = generateID()
	}

	if res.NameString != "" && res.Canonical == "" {
		parseNameString(gnp, &res)
	}
	return res
}

func parseNameString(gnp gnparser.GNparser, inp *Input) {
	parsed := gnp.ParseName(inp.NameString)
	if !parsed.Parsed {
		return
	}

	if parsed.Canonical != nil {
		inp.Canonical = parsed.Canonical.Simple
	}

	if parsed.Authorship != nil {
		inp.NameAuthors = strings.Join(parsed.Authorship.Authors, " ")

		if inp.NameYear == "" {
			inp.NameYear = parsed.Authorship.Year
		}
	}
}

func generateID() string {
	return uuid.NewString()
}
