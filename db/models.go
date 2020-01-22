package db

import (
	"database/sql"
)

type Item struct {
	ID             uint   `gorm:"primary_key;auto_increment:false"`
	BarCode        string `gorm:"type:varchar(60);unique_index;not null"`
	Vol            string `gorm:"type:varchar(100)"`
	YearStart      sql.NullInt32
	YearEnd        sql.NullInt32
	TitleID        uint   `gorm:"not null"`
	TitleDOI       string `gorm:"type:varchar(100)"`
	TitleName      string `gorm:"type:varchar(255)"`
	TitleYearStart sql.NullInt32
	TitleYearEnd   sql.NullInt32
	TitleLang      string `gorm:"type:varchar(20)"`
	PathsTotal     uint
	AnimaliaNum    uint
	PlantaeNum     uint
	FungiNum       uint
	BacteriaNum    uint
	MajorKingdom   string `gorm:"type:varchar(100)"`
	KingdomPercent uint
	Context        string `gorm:"type:varchar(100)"`
}

type Page struct {
	ID      uint `gorm:"primary_key;auto_increment:false"`
	ItemID  uint `gorm:"index:item;not null"`
	FileNum uint `gorm:"not null"`
	PageNum sql.NullInt64
}

type Part struct {
	ID                 uint `gorm:"primary_key;auto_increment:false"`
	PageID             sql.NullInt32
	ItemID             sql.NullInt32
	Length             sql.NullInt32
	DOI                string `gorm:"type:varchar(100)"`
	ContributorName    string `gorm:"type:varchar(255)"`
	SequenceOrder      sql.NullInt32
	SegmentType        string        `gorm:"type:varchar(100)"`
	Title              string        `gorm:"type:text"`
	ContainerTitle     string        `gorm:"type:text"`
	PublicationDetails string        `gorm:"type:text"`
	Volume             string        `gorm:"type:varchar(100)"`
	Series             string        `gorm:"type:varchar(100)"`
	Issue              string        `gorm:"type:varchar(100)"`
	Date               string        `gorm:"type:varchar(100)"`
	Year               sql.NullInt32 `gorm:"index:year"`
	YearEnd            sql.NullInt32
	Month              sql.NullInt32
	Day                sql.NullInt32
	PageNumStart       sql.NullInt32
	PageNumEnd         sql.NullInt32
	Language           string `gorm:"type:varchar(20)"`
}

type NameString struct {
	ID                string `sql:"type:uuid;primary_key"`
	Name              string `gorm:"type:varchar(255)"`
	TaxonID           string `gorm:"type:varchar(100)"`
	MatchType         string `gorm:"type:varchar(100)"`
	EditDistance      uint
	StemEditDistance  uint
	MatchedName       string `gorm:"type:varchar(255)"`
	MatchedCanonical  string `gorm:"type:varchar(255);index:canonical"`
	CurrentName       string `gorm:"type:varchar(255)"`
	CurrentCanonical  string `gorm:"type:varchar(255);index:current_canonical"`
	Classification    string
	DataSourceId      sql.NullInt32
	DataSourceTitle   string `gorm:"type:varchar(255)"`
	DataSourcesNumber uint
	Curation          bool `gorm:"index:curation"`
	Occurences        uint
	Odds              float32
	Error             string `gorm:"type:varchar(255)"`
}

type PageNameString struct {
	PageID       uint
	NameStringID string `sql:"type:uuid;index:name_string"`
	OffsetStart  uint
	OffsetEnd    uint
	Odds         float64
}
