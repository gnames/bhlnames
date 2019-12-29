package db

import (
	"database/sql"

	"github.com/jinzhu/gorm"
)

type Item struct {
	gorm.Model
	BarCode        string `gorm:"unique_index;not null"`
	Vol            string
	YearStart      uint `gorm:"not null"`
	YearEnd        uint `gorm:"not null"`
	TitleID        uint `gorm:"not null"`
	TitleName      string
	TitleYearStart uint `gorm:"not null"`
	TitleYearEnd   uint `gorm:"not null"`
}

type Page struct {
	gorm.Model
	ItemID  uint `gorm:"not null"`
	FileNum uint `gorm:"not null"`
	PageNum sql.NullInt64
}

type Part struct {
	gorm.Model
	PageID uint `gorm:"not null"`
	ItemID uint `gorm:"not null"`
	Length sql.NullInt64
}

type PartDetails struct {
	gorm.Model
	PartID             uint `gorm:"not null"`
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
	Month              sql.NullInt64
	Day                sql.NullInt64
	PageNumStart       sql.NullInt64
	PageNumEnd         sql.NullInt64
	Language           string
}
