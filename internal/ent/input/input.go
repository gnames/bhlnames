package input

import (
	"strconv"
	"strings"

	"github.com/gnames/gnparser"
	"github.com/google/uuid"
)

// @Description Input is used to pass data to the BHLnames API. It contains
// @Description infromation about a name and a reference where the name was
// @Description mentioned. Reference can point to a name usage or a
// @Description nomenclatural event.
type Input struct {
	// ID is a unique identifier for the Input. It is optional and helps
	// to find Input data on the client side.
	ID string `json:"id" example:"a1b2c3d4"`

	// Name provides data about a scientific name. Information can be
	// provided by a name-string or be split into separate fields.
	Name `json:"name"`

	// Reference provides data about a reference where the name was
	// mentioned. Information can be provided by a reference-string or
	// be split into separate fields.
	Reference `json:"reference"`
}

// @Description Name provides data about a scientific name.
type Name struct {
	// NameString is a scientific name as a string. It might be enough to
	// provide only NameString without provided other fields.
	NameString string `json:"nameString,omitempty" example:"Canis lupus Linnaeus, 1758"`

	// Canonical is the canonical form of a name, meaning the name without
	// authorship or a year.
	Canonical string `json:"canonical,omitempty" example:"Canis lupus"`

	// NameAuthors is the authorship of a name.
	NameAuthors string `json:"authors,omitempty" example:"Linnaeus"`

	// NameYear is the year of publication for a name.
	NameYear int `json:"year,omitempty" example:"1758"`
}

// @Description Reference provides data about a reference where the name was
// @Description mentioned.
type Reference struct {
	// RefString is a reference as a string. It might be enough to
	// provide only RefString without provided other fields.
	RefString string `json:"refString,omitempty" example:"Linnaeus, C. 1758. Systema naturae per regna tria naturae, secundum classes, ordines, genera, species, cum characteribus, differentiis, synonymis, locis. Tomus I. Editio decima, reformata. Holmiae: impensis direct. Laurentii Salvii. i–ii, 1–824 pp."`

	// RefYear is the year of publication for a reference.
	RefYearStart int `json:"yearStart,omitempty" example:"1758"`

	// RefYear is the year of publication for a reference.
	RefYearEnd int `json:"yearEnd,omitempty" example:"1758"`

	// RefAuthors is the authorship of a reference.
	RefAuthors string `json:"authors,omitempty" example:"Linnaeus"`

	// Journal is the title of the journal where the reference was
	// published.
	Journal string `json:"journal,omitempty" example:"Systema naturae per regna tria naturae, secundum classes, ordines, genera, species, cum characteribus, differentiis, synonymis, locis."`

	// Volume is the volume of the journal where the reference was
	// published.
	Volume int `json:"volume,omitempty" example:"1"`

	// PageStart is the first page of the reference.
	PageStart int `json:"pageStart,omitempty" example:"24"`

	// PageEnd is the last page of the reference.
	PageEnd int `json:"pageEnd,omitempty" example:"24"`
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
