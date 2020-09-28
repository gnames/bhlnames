package entity

type RefsResult struct {
	Output *Output
	Error  error
}

type Output struct {
	NameString       string       `json:"nameString"`
	Canonical        string       `json:"canonical,omitempty"`
	CurrentCanonical string       `json:"currentCanonical,omitempty"`
	Synonyms         []string     `json:"synonyms,omitempty"`
	ImagesUrl        string       `json:"imagesURL,omitempty"`
	ReferenceNumber  int          `json:"refsNum"`
	References       []*Reference `json:"references,omitempty"`
}

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
