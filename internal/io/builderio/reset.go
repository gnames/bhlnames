package builderio

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/gnames/bhlnames/internal/io/dbio"
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
	slog.Info("Creating tables.")
	err = model.Migrate(b.grm)
	if err != nil {
		slog.Error("Cannot create tables.", "err", err)
		return err
	}

	return nil
}

func (b builderio) resetDirs() error {
	err := gnsys.CleanDir(b.cfg.InputDir)
	if err != nil {
		slog.Error("Cannot clean input directory.", "err", err)
		return err
	}
	err = os.MkdirAll(b.cfg.DownloadDir, 0755)
	if err != nil {
		slog.Error("Cannot create download directory.", "err", err)
		return err
	}
	err = os.MkdirAll(b.cfg.AhoCorasickDir, 0755)
	if err != nil {
		slog.Error("Cannot create AhoCorasick directory.", "err", err)
		return err
	}
	err = dbio.ResetKeyVal(b.cfg.AhoCorKeyValDir)
	if err != nil {
		slog.Error("Cannot reset AhoCorasick key-value store.", "err", err)
		return err
	}
	return nil
}
