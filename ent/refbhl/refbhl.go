package refbhl

// ReferenceBHL is a representation of a BHL entity that was matched with a
// scientific name-string.
type ReferenceBHL struct {
	// YearAggr is the the most precise available year information for the
	// reference. It can come from reference year (part), year of a Volume
	// (item) or from title (usually book or journal).
	YearAggr int `json:"yearAggr"`
	// YearType indicates what got inserted as an YearAggr.
	YearType string `json:"yearType"`
	// URL is the URL of the reference in BHL.
	URL string `json:"url,omitempty"`
	// TitleDOI provides DOI for a book or journal
	TitleDOI string `json:"doiTitle,omitempty"`
	// PartDOI provides DOI for a part (usually a paper/publication).
	PartDOI string `json:"doiPart,omitempty"`
	// Name is a scientific name from the query.
	Name string `json:"name"`
	// MatchedName is a scientific name match from the reference's text.
	MatchName string `json:"matchName"`
	// EditDistance is the number of differences (edit events)
	// between Name and MatchName according to Levenshtein algorithm.
	EditDistance int `json:"editDistance,omitempty"`
	// AnnotNomen is a nomenclatural annotation located near the matchted name.
	AnnotNomen string `json:"annotNomen,omitempty"`
	// PageID is the BHL database ID for the page where the name was found.
	PageID int `json:"pageId"`
	// ItemID is the BHL database ID for Item (volume usually).
	ItemID int `json:"itemId"`
	// TitleID is the BHL database ID for the Title (book or journal).
	TitleID int `json:"titleId"`
	// PartID is the BHL database ID for the Part (usually a scientific paper).
	PartID int `json:"partId,omitempty"`
	// TitleName is the name of a title (a book or a journal).
	TitleName string `json:"titleName"`
	// TitleAbbr1 is normalized abbreviated title.
	TitleAbbr1 []string `json:"-"`
	// TitleAbbr2 is furhter abbreviated by removal of words that
	// are often ommited from lexical variants of a title.
	TitleAbbr2 []string `json:"-"`
	// Volume is the information about a volume in a journal.
	Volume string `json:"volume,omitempty"`
	// PartPages are the start and end pages of a publication.
	PartPages string `json:"partPages,omitempty"`
	// PartName is the publication title.
	PartName string `json:"partName,omitempty"`
	// ItemKingdom a consensus kingdom for names from the Item (journal volume).
	ItemKingdom string `json:"itemKingdom"`
	// ItemKingdomPercent, is the percentage showing how how many names belong to
	// the consensus Kingdom.
	ItemKingdomPercent int `json:"itemKingdomPercent"`
	// StatNamesNum is the number of names in the Item.
	StatNamesNum int `json:"statNamesNum"`
	// ItemContext provides a clade that contains a majority of scientific names
	// mentioned in the Item.
	ItemContext string `json:"itemContext"`
	// TitleYearStart is the year when book is published, or when a journal was
	// published first time.
	TitleYearStart int `json:"titleYearStart"`
	// TitleYearEnd is the year when a journal stopped being published.
	TitleYearEnd int `json:"titleYearEnd,omitempty"`
	// ItemYearStart is the year when an Item started to be published (most
	// items will have only ItemYearStart).
	ItemYearStart int `json:"itemYearStart,omitempty"`
	// ItemYearEnd is the year when an Item ended.
	ItemYearEnd int `json:"itemYearEnd,omitempty"`
	// Score is the oval score of matching of the reference with a name-string or
	// a reference-string.
	Score Score `json:"score"`
}

// Score gives a qualitative estimation of a quality of a match to a
// name-string, a nomen, or a reference-string.
type Score struct {
	// Sort is a value used to sort scores from best to worst. It is created as a
	// combination of all individual scores.
	Sort uint32 `json:"sortval"`

	// Total is a simple sum of all available individual score.
	Total int `json:"overal"`

	// Annot is a score important for nomenclatural events and provides match
	// for nomenclatural annotations.
	Annot int `json:"annot"`

	// Year is a score of a quality of a year match in a reference-string or
	// name-string.
	Year int `json:"year"`

	// RefTitle is the score of matching references titleName.
	RefTitle int `json:"refTitle"`
}
