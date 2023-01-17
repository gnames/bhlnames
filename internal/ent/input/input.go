package input

import (
	"strconv"
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
	NameYear    int    `json:"year,omitempty"`
	Canonical   string `json:"canonical,omitempty"`
	NameAuthors string `json:"authors,omitempty"`
}

type Reference struct {
	RefString    string `json:"refString,omitempty"`
	RefYearStart int    `json:"yearStart,omitempty"`
	RefYearEnd   int    `json:"yearEnd,omitempty"`
	RefAuthors   string `json:"authors,omitempty"`
	Journal      string `json:"journal,omitempty"`
	Volume       int    `json:"volume,omitempty"`
	PageStart    int    `json:"pageStart,omitempty"`
	PageEnd      int    `json:"pageEnd,omitempty"`
}

type Option func(*Input)

func OptID(s string) Option {
	return func(inp *Input) {
		inp.ID = s
	}
}

func OptNameString(s string) Option {
	return func(inp *Input) {
		inp.NameString = s
	}
}

func OptNameYear(i int) Option {
	return func(inp *Input) {
		inp.NameYear = i
	}
}

func OptRefString(s string) Option {
	return func(inp *Input) {
		inp.RefString = s
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
	if res.RefString != "" {
		parseRefString(&res)
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

		if inp.NameYear == 0 && parsed.Authorship.Year != "" {
			yr, _ := strconv.Atoi(parsed.Authorship.Year)
			inp.NameYear = yr
		}
	}
}

func generateID() string {
	return uuid.NewString()
}
