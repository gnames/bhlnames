package colbuildio

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/gnames/gnsys"
)

func (c colbuildio) deleteFiles() error {
	var err error
	var exists bool
	for _, v := range []string{c.pathDownload, c.pathExtract} {
		slog.Info("Removing file", "file", v)
		exists, err = gnsys.FileExists(v)
		if exists && err == nil {
			err = os.Remove(v)
		}
		if err != nil {
			err = fmt.Errorf("deleteFiles: %w", err)
			slog.Error("Cannot get info", "file", v, "error", err)
			return err
		}
	}
	return nil
}

func (c colbuildio) resetColDB() error {
	slog.Info("Rebuilding CoL database tables.")
	q := `DROP TABLE IF EXISTS `
	tables := []string{"col_nomen_refs", "col_bhl_refs"}
	for _, v := range tables {
		_, err := c.db.Exec(q + v)
		if err != nil {
			err = fmt.Errorf("resetColDB: %w", err)
			slog.Error("Cannot drop table", "table", v, "error", err)
			return err
		}
	}
	slog.Info("Recreating CoL tables.")
	c.gormDB.AutoMigrate(
		&db.ColNomenRef{},
		&db.ColBhlRefs{},
	)
	return nil
}
