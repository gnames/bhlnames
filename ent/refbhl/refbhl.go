package refbhl

// ReferenceBHL
type ReferenceBHL struct {
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
	Score              Score  `json:"score"`
}

type Score struct {
	Sort  uint32 `json:"-"`
	Total int    `json:"overal"`
	Annot int    `json:"annot"`
	Year  int    `json:"year"`
}
