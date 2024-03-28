package builderio

import (
	"log/slog"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/gnames/bhlnames/internal/ent/builder"
	"github.com/gnames/bhlnames/internal/io/bhlsys"
	"github.com/gnames/bhlnames/internal/io/dbio"
	"github.com/gnames/bhlnames/internal/io/namesbhlio"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
)

type builderio struct {
	cfg config.Config
	db  *pgxpool.Pool
	grm *gorm.DB
}

// New creates a new instance of the Builder and sets up necessary
// connections to the database.
func New(cfg config.Config) (builder.Builder, error) {
	var err error
	var db *pgxpool.Pool
	var grm *gorm.DB

	res := builderio{cfg: cfg}
	db, err = dbio.NewDB(cfg)
	if err != nil {
		return nil, err
	}

	grm, err = dbio.NewGORM(cfg)
	if err != nil {
		return nil, err
	}

	res.db = db
	res.grm = grm
	return &res, nil
}

func (b *builderio) ResetData() error {
	var err error
	slog.Info("Resetting filesystem.", "input-dir", b.cfg.InputDir)
	err = b.resetDirs()
	if err != nil {
		slog.Error("Cannot reset dirs fully.", "err", err)
		return err
	}

	err = b.resetDB()
	if err != nil {
		return err
	}

	return nil
}

func (b *builderio) ImportData() error {
	// Download and Extract
	err := b.downloadAndExtract()
	if err != nil {
		return err
	}

	// Reset Database
	err = b.resetDB()
	if err != nil {
		return err
	}

	// Import data coming from BHL dump
	err = b.importDataBHL()
	if err != nil {
		return err
	}

	// bloom filter is used to check if a name-string is already in the database
	var blf *bloom.BloomFilter
	n := namesbhlio.New(b.cfg, b.db, b.grm)

	// Import names coming from BHL index
	blf, err = n.ImportNames()
	if err != nil {
		return err
	}

	// Import occurrences of name-strings
	err = n.ImportOccurrences(blf)
	if err != nil {
		return err
	}

	return nil
}

func (b *builderio) Close() {
	b.db.Close()
	db, _ := b.grm.DB()
	db.Close()
}

func (b *builderio) downloadAndExtract() error {
	slog.Info(
		"Downloading database dump from BHL.",
		"url", b.cfg.BHLDumpURL,
		"file", b.cfg.DownloadBHLFile,
	)
	err := bhlsys.Download(
		b.cfg.DownloadBHLFile, b.cfg.BHLDumpURL, b.cfg.WithRebuild,
	)
	if err != nil {
		slog.Error("Cannot download BHL data.", "error", err)
		return err
	}

	slog.Info(
		"Extracting BHL database dump data.",
		"file", b.cfg.DownloadBHLFile,
	)
	err = bhlsys.Extract(
		b.cfg.DownloadBHLFile, b.cfg.DownloadDir, b.cfg.WithRebuild,
	)
	if err != nil {
		slog.Error("Cannot extract BHL data.",
			"file", b.cfg.DownloadBHLFile, "error", err,
		)
		return err
	}

	slog.Info(
		"Downloading names data from bhlindex.",
		"url", b.cfg.BHLNamesURL,
		"file", b.cfg.DownloadNamesFile,
	)
	err = bhlsys.Download(b.cfg.DownloadNamesFile, b.cfg.BHLNamesURL, b.cfg.WithRebuild)
	if err != nil {
		slog.Error("Cannot download names data.", "error", err)
		return err
	}

	slog.Info(
		"Extracting names data from bhlindex.",
		"file", b.cfg.DownloadNamesFile,
	)
	err = bhlsys.Extract(
		b.cfg.DownloadNamesFile, b.cfg.DownloadDir, b.cfg.WithRebuild,
	)
	if err != nil {
		slog.Error("Cannot extract names data.",
			"file", b.cfg.DownloadNamesFile, "error", err)
		return err
	}
	return nil
}
