package builderio

import (
	"database/sql"
	"fmt"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/gnames/bhlnames/internal/ent/builder"
	"github.com/gnames/bhlnames/internal/io/bhlsys"
	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/gnames/bhlnames/internal/io/namesbhlio"
	"github.com/gnames/bhlnames/pkg/config"
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

// Close closes all resources used by the Builder.
func (b builderio) Close() {
	b.DB.Close()
	b.GormDB.Close()
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

func (b builderio) PrepareData() error {
	var err error
	log.Info().Msg("Preparing data for bhlnames service.")
	return err
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
	var blf *bloom.BloomFilter
	n := namesbhlio.New(b.Config, b.DB, b.GormDB)

	// Download and Extract
	log.Info().Msg("Downloading database dump from BHL.")
	err := bhlsys.Download(b.DownloadBHLFile, b.BHLDumpURL, b.WithRebuild)
	if err == nil {
		err = bhlsys.Extract(b.DownloadBHLFile, b.DownloadDir, b.WithRebuild)
	}

	if err == nil {
		log.Info().Msg("Downloading names data from bhlindex.")
		err = bhlsys.Download(b.DownloadNamesFile, b.BHLNamesURL, b.WithRebuild)
	}
	if err == nil {
		err = bhlsys.Extract(b.DownloadNamesFile, b.DownloadDir, b.WithRebuild)
	}

	// Reset Database and Import Data
	if err == nil {
		b.resetDB()
		err = b.importDataBHL()
		if err != nil {
			err = fmt.Errorf("importDataBHL: %w", err)
		}
	}
	if err == nil {
		blf, err = n.ImportNames()
		if err != nil {
			err = fmt.Errorf("ImportNames: %w", err)
		}
	}

	if err == nil {
		err = n.ImportOccurrences(blf)
		if err != nil {
			err = fmt.Errorf("ImportOccurrences: %w", err)
		}
	}
	return err
}
