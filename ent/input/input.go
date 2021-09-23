package input

type Input struct {
	ID        string `json:"id"`
	Name      `json:"name"`
	Reference `json:"reference"`
}

type Name struct {
	NameString string `json:"nameString,omitempty"`
	NameYear   string `json:"year,omitempty"`
	Canonical  string `json:"canonical,omitempty"`
	Authorship string `json:"authorship,omitempty"`
}

type Reference struct {
	RefString string `json:"refString,omitempty"`
	RefYear   string `json:"year,omitempty"`
	Authors   string `json:"authors,omitempty"`
	Journal   string `json:"journal,omitempty"`
	Volume    string `json:"volume,omitempty"`
	Pages     string `json:"pages,omitempty"`
}
