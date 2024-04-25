package builderio

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/gnames/gnsys"
)

func (b builderio) resetDB() error {
	slog.Info("Resetting database.", "database",
		b.cfg.DbDatabase, "host", b.cfg.DbHost,
	)
	q := `
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO %s;
COMMENT ON SCHEMA public IS 'standard public schema'`
	q = fmt.Sprintf(q, b.cfg.DbUser)
	_, err := b.db.Exec(context.Background(), q)
	if err != nil {
		slog.Error("Cannot reset database.", "err", err)
		return err
	}

	slog.Info("Update collation.")
	q = `
UPDATE pg_database 
    SET datcollate = 'C', datctype = 'en_US.UTF-8' 
    WHERE datname = '%s';
`
	q = fmt.Sprintf(q, b.cfg.DbDatabase)
	_, err = b.db.Exec(context.Background(), q)
	if err != nil {
		slog.Error("Cannot update database's collation.", "err", err)
		return err
	}

	slog.Info("Creating tables.")
	err = model.Migrate(b.grm)
	if err != nil {
		slog.Error("Cannot create tables.", "err", err)
		return err
	}

	return nil
}

func (b builderio) resetDirs() error {
	var err error
	err = gnsys.MakeDir(b.cfg.RootDir)
	if err != nil {
		slog.Error(
			"Cannot create root directory",
			"dir", b.cfg.RootDir,
			"error", err,
		)
		return err
	}
	err = gnsys.CleanDir(b.cfg.RootDir)
	if err != nil {
		slog.Warn("Cannot clean input directory.", "err", err)
		slog.Info("Trying to create input directory.")
	}
	err = os.MkdirAll(b.cfg.ExtractDir, 0755)
	if err != nil {
		slog.Error("Cannot create download directory.", "err", err)
		return err
	}
	return nil
}
