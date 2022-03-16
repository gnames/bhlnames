package builderio

import (
	"database/sql"
	"log"

	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/builder"
	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/bhlnames/io/namesbhlio"
	"github.com/gnames/gnsys"
	"github.com/jinzhu/gorm"
)

type builderio struct {
	config.Config
	DB     *sql.DB
	GormDB *gorm.DB
}

func New(cfg config.Config) builder.Builder {
	res := builderio{
		Config: cfg,
		DB:     db.NewDB(cfg),
		GormDB: db.NewDbGorm(cfg),
	}
	res.touchDirs()
	return res
}

func (b builderio) touchDirs() {
	dirs := []string{
		b.InputDir,
		b.DownloadDir,
		b.PageDir,
		b.PageFileDir,
		b.PartDir,
		b.AhoCorasickDir,
		b.AhoCorKeyValDir,
	}
	for i := range dirs {
		exists, _, _ := gnsys.DirExists(dirs[i])
		if !exists {
			err := gnsys.MakeDir(dirs[i])
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (b builderio) ResetData() {
	var err error

	log.Printf("Reseting filesystem at '%s'.", b.InputDir)
	err = b.resetDirs()
	if err != nil {
		log.Fatalf("Cannot reset dirs: %s.", err)
	}
	b.resetDB()
}

func (b builderio) ImportData() error {
	n := namesbhlio.New(b.Config, b.DB, b.GormDB)
	err := b.downloadDumpBHL()

	if err == nil {
		err = b.extractFilesBHL()
	}

	if err == nil {
		b.resetDB()
		err = b.importDataBHL()
	}

	if err == nil {
		err = n.ImportNames()
	}

	if err == nil {
		err = n.PageFilesToIDs()
	}

	if err == nil {
		err = n.ImportOccurrences()
	}

	return err
}
