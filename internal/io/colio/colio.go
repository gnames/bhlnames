package colio

import (
	"log/slog"
	"path/filepath"

	"github.com/gnames/bhlnames/internal/ent/col"
	"github.com/gnames/bhlnames/internal/io/bhlsys"
	"github.com/gnames/bhlnames/internal/io/dbio"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnsys"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
)

type colio struct {
	cfg config.Config
	db  *pgxpool.Pool
	grm *gorm.DB

	gnpPool chan gnparser.GNparser

	recordsNum  int
	lastProcRec int
}

// New creates CoL's Nomen instance
func New(cfg config.Config) (col.Nomen, error) {
	db, err := dbio.NewDB(cfg)
	if err != nil {
		return nil, err
	}

	grm, err := dbio.NewGORM(cfg)
	if err != nil {
		return nil, err
	}

	gnpPool := gnparser.NewPool(gnparser.NewConfig(), cfg.JobsNum)

	res := colio{
		cfg:     cfg,
		db:      db,
		grm:     grm,
		gnpPool: gnpPool,
	}
	return &res, nil
}

// CheckCoLData verifies if the CoL archive file exists and if its contents
// have been previously extracted. Returns flags indicating their status.
// It also checks if there are CoL-related records in the database.
func (c *colio) CheckCoLData() (bool, bool, error) {
	var err error
	var exists, hasFiles, hasData bool

	exists, _ = gnsys.FileExists(c.cfg.DownloadCoLFile)
	if exists {
		pathExtract := filepath.Join(c.cfg.ExtractDir, "Taxon.tsv")
		hasFiles, _ = gnsys.FileExists(pathExtract)
	}

	hasData, err = c.checkData()
	return hasFiles, hasData, err
}

// ResetCoLData removes all CoL-related data (downloaded files, generated
// resources). This restores the system to a clean state with no CoL data.
func (c *colio) ResetCoLData() error {
	slog.Info("Reseting CoL files.")
	err := c.deleteFiles()
	if err != nil {
		slog.Error("Cannot delete files", "error", err)
		return err
	}
	err = c.resetColDB()
	if err != nil {
		slog.Error("Cannot reset CoL database tables", "error", err)
		return err
	}
	return nil
}

// ImportCoLData downloads the CoL Darwin Core Archive and imports relevant
// taxonomic names and references into the internal storage.
func (c *colio) ImportCoLData() error {
	var err error
	slog.Info("Downloading CoL DwCA data.")
	err = bhlsys.Download(c.cfg.DownloadCoLFile, c.cfg.CoLDataURL, false)
	if err != nil {
		slog.Error("Cannot download CoL data", "error", err)
		return err
	}

	err = bhlsys.Extract(c.cfg.DownloadCoLFile, c.cfg.ExtractDir, false)
	if err != nil {
		slog.Error("Cannot extract CoL data", "error", err)
		return err
	}

	err = c.importCoL()
	if err != nil {
		slog.Error("Cannot import CoL data", "error", err)
		return err
	}
	return nil
}

// Close releases all resources (e.g., database connections) used by the
// Nomen instance.
func (c *colio) Close() {
	c.db.Close()
	db, _ := c.grm.DB()
	db.Close()
}
