package bhlnames

import (
	"github.com/gnames/bhlnames/db"
	"github.com/jinzhu/gorm"
)

func (bhln BHLnames) Init() error {
	d, err := db.NewDb(bhln.DbHost, bhln.DbUser, bhln.DbPass, bhln.DbName)
	if err != nil {
		return err
	}
	migrate(d)
	defer d.Close()
	return nil
}

func migrate(d *gorm.DB) {
	d.AutoMigrate(
		&db.Item{},
		&db.Page{},
		&db.Part{},
		&db.PartDetails{},
	)
}
