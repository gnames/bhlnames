package bhlnames

import (
	"fmt"
	"log"

	"github.com/gnames/bhlnames/db"
	"github.com/gnames/bhlnames/names"
	"github.com/gnames/bhlnames/sys"
	"github.com/jinzhu/gorm"
)

func (bhln BHLnames) Init() error {
	err := bhln.BHL()
	if err != nil {
		return err
	}
	return bhln.Names()
}

func (bhln BHLnames) Names() error {
	log.Println("Populating database with names occurances data")
	n := names.NewNames(bhln.BHLindexHost, bhln.DbOpts, bhln.InputDir)
	err := n.ImportNames()
	if err != nil {
		return err
	}
	return n.ImportNamesOccur(bhln.KeyValDir)
}

func (bhln BHLnames) BHL() error {
	log.Printf("Run Migrations for Postgresql database '%s'", bhln.DbOpts.Name)
	d := bhln.DbOpts.NewDbGorm()
	migrate(d)
	d.Close()
	log.Println("Migrations done")
	return bhln.getMetadata()
}

func migrate(d *gorm.DB) {
	d.AutoMigrate(
		&db.Item{},
		&db.Page{},
		&db.Part{},
		&db.NameString{},
		&db.PageNameString{},
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
	fmt.Println()
	log.Println("BHL metadata is uploaded to db.")
	fmt.Println()
	return nil
}
