package builderio

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/gnames/gnsys"
)

func (b builderio) resetDB() error {
	slog.Info("Resetting database", "database", b.DbDatabase, "host", b.DbHost)
	q := `
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO %s;
COMMENT ON SCHEMA public IS 'standard public schema'`
	q = fmt.Sprintf(q, b.DbUser)
	_, err := b.DB.Exec(q)
	if err != nil {
		err = fmt.Errorf("builderio.resetDB: %w", err)
		slog.Error("Cannot reset database", "err", err)
		return err
	}
	slog.Info("Creating tables.")
	err = b.migrate()
	if err != nil {
		return err
	}

	return nil
}

func (b builderio) migrate() error {
	b.GormDB.AutoMigrate(
		&db.Item{},
		&db.ItemStats{},
		&db.Page{},
		&db.Part{},
		&db.NameString{},
		&db.NameOccurrence{},
		&db.ColNomenRef{},
		&db.ColBhlRefs{},
	)
	err := db.Truncate(b.DB, []string{"items", "pages", "parts"})
	if err != nil {
		err = fmt.Errorf("builderio.migrate: %w", err)
		slog.Error("Cannot truncate tables", "err", err)
		return err
	}
	return nil
}

func (b builderio) resetDirs() error {
	err := gnsys.CleanDir(b.DownloadDir)
	if err != nil {
		return err
	}
	err = db.ResetKeyVal(b.PageDir)
	if err != nil {
		return err
	}
	err = db.ResetKeyVal(b.PartDir)
	if err != nil {
		return err
	}
	err = db.ResetKeyVal(b.AhoCorKeyValDir)
	if err != nil {
		return err
	}
	exists, _ := gnsys.FileExists(b.DownloadBHLFile)
	if exists {
		return os.Remove(b.DownloadBHLFile)
	}
	return nil
}
