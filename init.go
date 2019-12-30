package bhlnames

import (
	"log"

	"github.com/gnames/bhlnames/db"
	"github.com/gnames/bhlnames/sys"
	"github.com/jinzhu/gorm"
)

func (bhln BHLnames) Init() error {
	log.Printf("Run Migrations for Postgresql database '%s'", bhln.DbOpts.Name)
	d := bhln.DbOpts.NewDbGorm()
	migrate(d)
	d.Close()
	log.Println("Migrations done")
	err := bhln.getMetadata()
	if err != nil {
		return err
	}
	return nil
}

func migrate(d *gorm.DB) {
	d.AutoMigrate(
		&db.Item{},
		&db.Kingdoms{},
		&db.Page{},
		&db.Part{},
	)
	db.Truncate(d)
}

func (bhln BHLnames) getMetadata() error {
	md := bhln.MetaData
	err := sys.MakeDir(bhln.InputDir)
	if err != nil {
		return nil
	}
	err = md.Download()
	if err != nil {
		return err
	}
	err = md.Extract()
	if err != nil {
		return err
	}
	err = md.Prepare()
	if err != nil {
		return err
	}
	return nil
}
