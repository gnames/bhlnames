package bhl

import "github.com/gnames/bhlnames/internal/ent/input"

// @Description RefsByName provides references to BHL Items, Parts and Pages
// @Description where a name-string, taxon or a putative nomenclatural
// @Description event were found.
type RefsByName struct {
	// Meta provides metadata for the results of a name-string search.
	Meta `json:"meta"`
	// References is a list of references to BHL Items, Parts and Pages
	// where a name-string was found.
	References []*ReferenceName `json:"references,omitempty"`
}

// @Description Meta provides metadata for the results of a name-string search.
type Meta struct {
	// Input of a name and/or reference
	Input input.Input `json:"input"`

	// NomenEventFromCache indicates that nomenclatural event was taken from
	// a pre-cached data.
	NomenEventFromCache bool `json:"nomenEventFromCache,omitempty"`

	// InputReferenceFrom indicates that input references were taked from
	// a data source.
	InputReferenceFrom string `json:"inputReferenceFrom,omitempty"`

	// Canonical is a full canonical form of the input name-string.
	Canonical string `json:"canonical,omitempty"`

	// CurrentCanonical is a full canonical form of a currently accepted
	// name for the taxon of the input name-string.
	CurrentCanonical string `json:"currentCanonical,omitempty"`

	// Synonyms is a list of synonyms for the name-string.
	Synonyms []string `json:"synonyms,omitempty"`

	// Error in results
	Error error `json:"error,omitempty"`

	// ReferenceNumber is the number of references found for the name-string.
	ReferenceNumber int `json:"totalRefsNum,omitempty"`
}
