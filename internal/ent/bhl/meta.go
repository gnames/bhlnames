package bhl

import "github.com/gnames/bhlnames/internal/ent/input"

// RefsByName provides apparent occurrences of a name-string in BHL.
type RefsByName struct {
	*Meta
	References []*ReferenceName `json:"references,omitempty"`
}

type Meta struct {
	// Input of a name and/or reference
	Input input.Input `json:"input"`

	// Canonical is a full canonical form of the input name-string.
	Canonical string `json:"canonical,omitempty"`

	// CurrentCanonical is a full canonical form of a currently accepted
	// name for the taxon of the input name-string.
	CurrentCanonical string `json:"currentCanonical,omitempty"`

	// Synonyms is a list of synonyms for the name-string.
	Synonyms []string `json:"synonyms,omitempty"`

	// ImagesURL provides URL that contains images of the taxon.
	ImagesURL string `json:"imagesURL,omitempty"`

	// Error in results
	Error error `json:"error,omitempty"`

	// ReferenceNumber is the number of references found for the name-string.
	ReferenceNumber int `json:"totalRefsNum,omitempty"`
}
