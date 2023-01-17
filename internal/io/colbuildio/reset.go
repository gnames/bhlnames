package colbuildio

import (
	"fmt"
	"os"

	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/gnames/gnsys"
	"github.com/rs/zerolog/log"
)

func (c colbuildio) deleteFiles() {
	var err error
	var exists bool
	for _, v := range []string{c.pathDownload, c.pathExtract} {
		log.Info().Msgf("Removing file '%s'.", v)
		exists, err = gnsys.FileExists(v)
		if exists && err == nil {
			err = os.Remove(v)
		}
		if err != nil {
			err = fmt.Errorf("deleteFiles: %w", err)
			log.Fatal().Err(err).Msg("")
		}
	}

}

func (c colbuildio) resetColDB() {
	log.Info().Msg("Rebuilding CoL database tables.")
	q := `DROP TABLE IF EXISTS `
	tables := []string{"col_nomen_refs", "col_bhl_refs"}
	for _, v := range tables {
		_, err := c.db.Exec(q + v)
		if err != nil {
			err = fmt.Errorf("resetColDB: %w", err)
			log.Fatal().Err(err).Msg("")
		}
	}
	log.Info().Msg("Recreating CoL tables.")
	c.gormDB.AutoMigrate(
		&db.ColNomenRef{},
		&db.ColBhlRefs{},
	)
}
