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
	Class          string
	Context        string
}

type Kingdoms struct {
	ItemID   uint
	Animalia uint
	Plantae  uint
	Fungi    uint
	Bacteria uint
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
