package namerefs

import (
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/refbhl"
)

// NameRefs provides apparent occurrences of a name-string in BHL.
type NameRefs struct {
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

	// ReferenceNumber is the number of references found for the name-string.
	ReferenceNumber int `json:"refsNum"`

	// References is a list of all unique BHL references to the name occurence.
	References []*refbhl.ReferenceBHL `json:"references,omitempty"`

	// Error in the kk
	Error error

	// WithSynonyms sets an option of returning references for synonyms of a name
	// as well.
	WithSynonyms bool
}

// func (nr *NameRefs) DetectNomen() {
// 	for i := range nr.References {
// 		nr.References[i].GetNomenScore()
// 	}
// 	sort.Slice(nr.References, func(i, j int) bool {
// 		r1 := nr.References[i]
// 		r2 := nr.References[j]
// 		return r1.Score.Overall() > r2.Score.Overall()
// 	})
// }
