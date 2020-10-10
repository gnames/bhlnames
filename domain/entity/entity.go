package entity

import "github.com/gnames/bhlnames/config"

// NameRefs provides apparent occurrences of a name-string in BHL.
type NameRefs struct {
	// NameString is the input name-string (verbatim).
	NameString string `json:"nameString"`
	// Canonical is a full canonical form of the input name-string.
	Canonical string `json:"canonical,omitempty"`
	// CurrentCanonical is a full canonical form of a currently accepted
	// name for the taxon of the input name-string.
	CurrentCanonical string `json:"currentCanonical,omitempty"`
	// Synonyms is a list of synonyms for the name-string.
	Synonyms []string `json:"synonyms,omitempty"`
	// ImagesURL provides URL that contains images of the taxon.
	ImagesUrl string `json:"imagesURL,omitempty"`
	// ReferenceNumber is the number of references found for the name-string.
	ReferenceNumber int `json:"refsNum"`
	// References is a list of all unique BHL references to the name occurence.
	References []*Reference `json:"references,omitempty"`
	// Parameters are settings of the query
	Params config.Search
}

// Reference
type Reference struct {
	YearAggr           int    `json:"yearAggr"`
	YearType           string `json:"yearType"`
	URL                string `json:"url,omitempty"`
	TitleDOI           string `json:"doiTitle,omitempty"`
	PartDOI            string `json:"doiPart,omitempty"`
	Name               string `json:"name"`
	MatchName          string `json:"matchName"`
	EditDistance       int    `json:"editDistance,omitempty"`
	AnnotNomen         string `json:"annotNomen,omitempty"`
	PageID             int    `json:"pageId"`
	ItemID             int    `json:"itemId"`
	TitleID            int    `json:"titleId"`
	PartID             int    `json:"partId,omitempty"`
	TitleName          string `json:"titleName"`
	Volume             string `json:"volume,omitempty"`
	PartPages          string `json:"partPages,omitempty"`
	PartName           string `json:"partName,omitempty"`
	ItemKingdom        string `json:"itemKingdom"`
	ItemKingdomPercent int    `json:"itemKingdomPercent"`
	StatNamesNum       int    `json:"statNamesNum"`
	ItemContext        string `json:"itemContext"`
	TitleYearStart     int    `json:"titleYearStart"`
	TitleYearEnd       int    `json:"titleYearEnd,omitempty"`
	ItemYearStart      int    `json:"itemYearStart,omitempty"`
	ItemYearEnd        int    `json:"itemYearEnd,omitempty"`
}
