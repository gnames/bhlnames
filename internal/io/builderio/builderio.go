package builderio

import (
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/gnames/bhlnames/internal/ent/builder"
	"github.com/gnames/bhlnames/internal/io/bhlsys"
	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/gnames/bhlnames/internal/io/namesbhlio"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnsys"
	"github.com/jinzhu/gorm"
)

type builderio struct {
	config.Config
	DB     *sql.DB
	GormDB *gorm.DB
}

func New(cfg config.Config) (builder.Builder, error) {
	dbConn, err := db.NewDB(cfg)
	if err != nil {
		return nil, err
	}
	gormDB, err := db.NewDbGorm(cfg)
	if err != nil {
		return nil, err
	}
	res := builderio{
		Config: cfg,
		DB:     dbConn,
		GormDB: gormDB,
	}
	err = res.touchDirs()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Close closes all resources used by the Builder.
func (b builderio) Close() {
	b.DB.Close()
	b.GormDB.Close()
}

func (b builderio) touchDirs() error {
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
				slog.Error("Cannot create dir", "dir", dirs[i], "err", err)
				return err
			}
		}
	}
	return nil
}

func (b builderio) PrepareData() error {
	var err error
	path := filepath.Join(b.AhoCorasickDir, "patterns.txt")
	exists, err := gnsys.FileExists(path)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	slog.Info("Preparing data for bhlnames service.")
	titlesMap, err := b.dbTitlesMap()
	if err != nil {
		return err
	}
	ts, err := newTitleStore(b.Config, titlesMap)
	if err != nil {
		return err
	}

	err = ts.setup()
	if err != nil {
		return err
	}

	return nil
}

func (b builderio) ResetData() error {
	var err error
	slog.Info("Resetting filesystem", "input-dir", b.InputDir)
	err = b.resetDirs()
	if err != nil {
		err = fmt.Errorf("builderio.ResetData: %#w", err)
		slog.Error("Cannot reset dirs", "err", err)
		return err
	}

	err = b.resetDB()
	if err != nil {
		return err
	}

	return nil
}

func (b builderio) ImportData() error {
	var blf *bloom.BloomFilter
	n := namesbhlio.New(b.Config, b.DB, b.GormDB)

	// Download and Extract
	slog.Info("Downloading database dump from BHL.")
	err := bhlsys.Download(b.DownloadBHLFile, b.BHLDumpURL, b.WithRebuild)
	if err == nil {
		err = bhlsys.Extract(b.DownloadBHLFile, b.DownloadDir, b.WithRebuild)
	}

	if err == nil {
		slog.Info("Downloading names data from bhlindex.")
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
