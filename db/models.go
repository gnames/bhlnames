package db

import (
	"database/sql"
)

type Item struct {
	ID             uint   `gorm:"primary_key;auto_increment:false"`
	BarCode        string `gorm:"unique_index;not null"`
	Vol            string
	YearStart      sql.NullInt64
	YearEnd        sql.NullInt64
	TitleID        uint `gorm:"not null"`
	TitleDOI       string
	TitleName      string
	TitleYearStart sql.NullInt64
	TitleYearEnd   sql.NullInt64
	TitleLang      string
	PathsTotal     uint
	AnimaliaNum    uint
	PlantaeNum     uint
	FungiNum       uint
	BacteriaNum    uint
	MajorKingdom   string
	KingdomPercent uint
	Context        string
}

type Page struct {
	ID      uint `gorm:"primary_key;auto_increment:false"`
	ItemID  uint `gorm:"index:item;not null"`
	FileNum uint `gorm:"not null"`
	PageNum sql.NullInt64
}

type Part struct {
	ID                 uint `gorm:"primary_key;auto_increment:false"`
	PageID             sql.NullInt64
	ItemID             sql.NullInt64
	Length             sql.NullInt64
	DOI                string
	ContributorName    string
	SequenceOrder      sql.NullInt64
	SegmentType        string
	Title              string
	ContainerTitle     string
	PublicationDetails string
	Volume             string
	Series             string
	Issue              string
	Date               string
	Year               sql.NullInt64 `gorm:"index:year"`
	YearEnd            sql.NullInt64
	Month              sql.NullInt64
	Day                sql.NullInt64
	PageNumStart       sql.NullInt64
	PageNumEnd         sql.NullInt64
	Language           string
}

type NameString struct {
	ID                string `sql:"type:uuid;primary_key"`
	Name              string
	TaxonID           string
	MatchType         string
	EditDistance      uint
	StemEditDistance  uint
	MatchedName       string
	MatchedCanonical  string `gorm:"index:canonical"`
	CurrentName       string
	CurrentCanonical  string `gorm:"index:current_canonical"`
	Classification    string
	DataSourceId      sql.NullInt64
	DataSourceTitle   string
	DataSourcesNumber uint
	Curation          bool `gorm:"index:curation"`
	Occurences        uint
	Odds              float32
	Error             string
}

type PageNameString struct {
	PageID       uint
	NameStringID string `sql:"type:uuid;index:name_string"`
	OffsetStart  uint
	OffsetEnd    uint
	Odds         float64
}
