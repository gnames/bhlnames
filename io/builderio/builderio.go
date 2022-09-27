package builderio

import (
	"database/sql"
	"fmt"

	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/builder"
	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/bhlnames/io/namesbhlio"
	"github.com/gnames/gnsys"
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog/log"
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
				err = fmt.Errorf("builderio.touchDirs: %#w", err)
				log.Fatal().Err(err).Msg("touchDirs")
			}
		}
	}
}

func (b builderio) ResetData() {
	var err error

	log.Info().Msgf("Reseting filesystem at '%s'.", b.InputDir)
	err = b.resetDirs()
	if err != nil {
		err = fmt.Errorf("builderio.ResetData: %#w", err)
		log.Fatal().Err(err).Msg("Cannot reset dirs")
	}
	b.resetDB()
}

func (b builderio) ImportData() error {
	n := namesbhlio.New(b.Config, b.DB, b.GormDB)

	log.Info().Msg("Downloading database dump from BHL.")
	err := b.download(b.DownloadBHLFile, b.BHLDumpURL)
	if err == nil {
		err = b.extract(b.DownloadBHLFile)
	}
	if err == nil {
		log.Info().Msg("Downloading names data from bhlindex.")
		err = b.download(b.DownloadNamesFile, b.BHLNamesURL)
	}
	if err == nil {
		err = b.extract(b.DownloadNamesFile)
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
