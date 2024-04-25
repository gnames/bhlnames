package colio

import (
	"context"
	"log/slog"
	"os"

	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/gnames/gnsys"
)

func (c colio) deleteFiles() error {
	var err error
	var exists bool
	for _, v := range []string{c.cfg.DownloadCoLFile} {
		slog.Info("Removing file", "file", v)
		exists, err = gnsys.FileExists(v)
		if err != nil {
			slog.Error("Cannot get file info", "file", v, "error", err)
			return err
		}
		if exists {
			err = os.Remove(v)
		}
		if err != nil {
			slog.Error("Cannot delete file", "file", v, "error", err)
			return err
		}
	}
	return nil
}

func (c colio) resetColDB() error {
	ctx := context.Background()
	slog.Info("Rebuilding CoL database tables.")
	q := `DROP TABLE IF EXISTS `
	tables := []string{"col_names", "col_bhl_refs"}

	for _, v := range tables {
		_, err := c.db.Exec(ctx, q+v)
		if err != nil {
			slog.Error("Cannot drop table", "table", v, "error", err)
			return err
		}
	}

	slog.Info("Recreating CoL tables.")
	c.grm.AutoMigrate(
		&model.ColName{},
		&model.ColBhlRef{},
	)
	return nil
}
