package builder_pg

import (
	"database/sql"
	"log"

	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/data/builder_pg/names"
	"github.com/gnames/bhlnames/db"
	"github.com/jinzhu/gorm"
)

type BuilderPG struct {
	config.Config
	DB     *sql.DB
	GormDB *gorm.DB
}

func NewBuilderPG(cfg config.Config) BuilderPG {
	res := BuilderPG{
		Config: cfg,
		DB:     db.NewDB(cfg.DB),
		GormDB: db.NewDbGorm(cfg.DB),
	}
	return res
}

func (b BuilderPG) ResetData() {
	log.Printf("Resetting '%s' database at '%s'.", b.Config.DB.Name, b.Config.DB.Host)
	b.resetDB()
	log.Print("Creating tables.")
	b.migrate()
	log.Printf("Reseting filesystem at '%s'.", b.Config.FileSystem.InputDir)
	err := b.resetDirs()
	if err != nil {
		log.Fatalf("Cannot reset dirs: %s.", err)
	}
}

func (b BuilderPG) ImportData() error {
	// err := b.downloadDumpBHL()
	// if err != nil {
	// 	return err
	// }
	// err = b.extractFilesBHL()
	// if err != nil {
	// 	return err
	// }
	// err = b.uploadDataBHL()
	// if err != nil {
	// 	return err
	// }

	log.Println("Populating database with names occurences data")
	n := names.NewNames(b.Config.BHLindexHost, b.Config.InputDir)
	n.DB = b.DB
	n.GormDB = b.GormDB
	err := n.ImportNames()
	if err != nil {
		return err
	}
	return n.ImportNamesOccur(b.Config.KeyValDir)
}
