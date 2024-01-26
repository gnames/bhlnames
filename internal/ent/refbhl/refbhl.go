package refbhl

import bout "github.com/gnames/bayes/ent/output"

// @Description ReferenceNameBHL represents a BHL entity that
// @Description includes a matched scientific name and the reference where
// @Description this name was discovered.
type ReferenceNameBHL struct {
	// Reference is the BHL reference where the name was detected.
	Reference `json:"reference"`

	// NameData contains detailed information about the scientific name.
	*NameData `json:"name,omitempty"`

	// IsNomenRef states is the reference likely contains
	// a nomenclatural event for the name.
	IsNomenRef bool `json:"isNomenRef" example:"true"`

	// Score is the overall score of the match between the reference and
	// a name-string or a reference-string.
	Score `json:"score"`
}

// @Description NameData contains details about a scientific name
// @Description provided in the search.
type NameData struct {
	// Name is a scientific name from the query.
	Name string `json:"name" example:"Pardosa moesta"`

	// MatchedName is a scientific name match from the reference's text.
	MatchedName string `json:"matchName" example:"Pardosa moesta Banks, 1892"`

	// EditDistance is the number of differences (edit events)
	// between Name and MatchName according to Levenshtein algorithm.
	EditDistance int `json:"editDistance,omitempty" example:"0"`

	// AnnotNomen is a nomenclatural annotation located near the matchted name.
	AnnotNomen string `json:"annotNomen,omitempty" example:"sp. nov."`
}

// @Description Reference represents a BHL reference that matched the query.
// @Description This could be a book, a journal, or a scientific paper.
type Reference struct {
	// YearAggr is the most precise year information available for the
	// reference. This could be from the reference year (part),
	// the year of a Volume (item), or from the title (usually a book
	// or journal).
	YearAggr int `json:"yearAggr" example:"1892"`

	// YearType indicates the source of the YearAggr value.
	YearType string `json:"yearType" example:"part"`

	// TitleID is the BHL database ID for the Title (book or journal).
	TitleID int `json:"titleId" example:"12345"`

	// TitleName is the name of a title (a book or a journal).
	TitleName string `json:"titleName" example:"Bulletin of the American Museum of Natural History"`

	// TitleAbbr1 is the normalized abbreviated title.
	TitleAbbr1 []string `json:"-"`

	// TitleAbbr2 is furhter abbreviated by removal of words that
	// are often ommitted from lexical variants of a title.
	TitleAbbr2 []string `json:"-"`

	// TitleDOI provides DOI for a book or journal
	TitleDOI string `json:"doiTitle,omitempty" example:"10.1234/5678"`

	// TitleYearStart is the year the when book is published,
	// or when the journal was first published.
	TitleYearStart int `json:"titleYearStart" example:"1890"`

	// TitleYearEnd is the year when the journal ceased publication.
	TitleYearEnd int `json:"titleYearEnd,omitempty" example:"1922"`

	// ItemID is the BHL database ID for Item (usually a volume).
	ItemID int `json:"itemId" example:"12345"`

	// Volume is the information about a volume in a journal.
	Volume string `json:"volume,omitempty" example:"vol. 12"`

	// ItemYearStart is the year when an Item began publication (most
	// items will have only ItemYearStart).
	ItemYearStart int `json:"itemYearStart,omitempty" example:"1892"`

	// ItemYearEnd is the year when an Item ceased publication.
	ItemYearEnd int `json:"itemYearEnd,omitempty" example:"1893"`

	// PageID is the BHL database ID for the page where the name was found.
	// It is provided by BHL.
	PageID int `json:"pageId" example:"12345"`

	// PageNum is the page number provided by the hard copy of the publication.
	PageNum int `json:"pageNum,omitempty" example:"123"`

	// URL is the URL of the reference in BHL.
	URL string `json:"url,omitempty" example:"https://www.biodiversitylibrary.org/page/12345"`

	// Part corresponds to a scientific paper, or other
	// distinct entity in an Item.
	*Part `json:"part,omitempty"`

	// ItemStats provides insights about the Reference Item.
	// From this data it is possible to infer what kind of
	// taxonomic groups are prevalent in the text.
	ItemStats `json:"itemStats"`
}

// @Description Part represents a distinct entity, usually a scientific paper,
// within an Item.
type Part struct {
	// ID is the BHL database ID for the Part (usually a scientific paper).
	ID int `json:"id,omitempty" example:"39371"`

	// Pages are the start and end pages of a publication.
	Pages string `json:"pages,omitempty" example:"925-928"`

	// Year is the year of publication for a part.
	Year int `json:"year,omitempty" example:"1886"`

	// Name is the publication title.
	Name string `json:"name,omitempty" example:"On a remarkable bacterium (Streptococcus) from wheat-ensilage"`

	// DOI provides DOI for a part (usually a paper/publication).
	DOI string `json:"doi,omitempty" example:"10.1234/5678"`
}

// @Description ItemStats provides insights about a Reference's Item.
// @Description This data can be used to infer the prevalent taxonomic
// @Description groups within the Item.
type ItemStats struct {
	// ItemKingdom a consensus kingdom for names from the Item (journal volume).
	ItemKingdom string `json:"itemKingdom" example:"Animalia"`

	// ItemKingdomPercent indicates the percentage of names that belong
	// to the consensus Kingdom.
	ItemKingdomPercent int `json:"itemKingdomPercent" example:"81"`

	// UniqNamesNum is the number of unique names in the Item.
	UniqNamesNum int `json:"uniqNamesNum" example:"1234"`

	// ItemMainTaxon provides a clade that contains a majority of scientific names
	// mentioned in the Item.
	ItemMainTaxon string `json:"itemMainTaxon" example:"Arthropoda"`
}

// @Description Score provides a qualitative estimation of a match quality
// @Description to a name-string, a nomen, or a reference-string.
type Score struct {
	//Odds is total Naive Bayes odds for the score.
	Odds float64 `json:"odds" example:"0.1234"`

	// OddsDetail provides details of the odds calculation.
	OddsDetail bout.OddsDetails `json:"oddsDetail"`

	// Total is a simple sum of all available individual scores.
	Total int `json:"total" example:"15"`

	// Annot is a score important for nomenclatural events and provides match
	// for nomenclatural annotations.
	Annot int `json:"annot" example:"3"`

	// Year is a score representing the quality of a year match
	// in a reference-string or the name-string.
	Year int `json:"year" example:"3"`

	// RefTitle is the score of matching reference's titleName.
	RefTitle int `json:"title" example:"3"`

	// RefVolume is a score derived from matching volume from
	// reference and BHL Volume.
	RefVolume int `json:"volume" example:"3"`

	// RefPages is a score derived from matching pages in a reference
	// and a page in BHL.
	RefPages int `json:"pages" example:"3"`

	// Labels provide types for each match
	Labels map[string]string `json:"labels"`
}
