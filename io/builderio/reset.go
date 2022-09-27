package builderio

import (
	"fmt"
	"os"

	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/gnsys"
	"github.com/rs/zerolog/log"
)

func (b builderio) resetDB() {
	log.Info().Msgf("Resetting '%s' database at '%s'.", b.DbDatabase, b.DbHost)
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
		log.Fatal().Err(err).Msg("Database reset failed")
	}
	log.Info().Msg("Creating tables.")
	b.migrate()
}

func (b builderio) migrate() {
	b.GormDB.AutoMigrate(
		&db.Item{},
		&db.ItemStats{},
		&db.Page{},
		&db.Part{},
		&db.NameString{},
		&db.NameOccurrence{},
	)
	err := db.Truncate(b.DB, []string{"items", "pages", "parts"})
	if err != nil {
		err = fmt.Errorf("builderio.migrate: %w", err)
		log.Fatal().Err(err).Msg("migration")
	}
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
