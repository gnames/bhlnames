package builderio

import (
	"fmt"
	"log"
	"os"

	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/gnsys"
)

func (b builderio) resetDB() {
	log.Printf("Resetting '%s' database at '%s'.", b.DbName, b.DbHost)
	q := `
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO %s;
COMMENT ON SCHEMA public IS 'standard public schema'`
	q = fmt.Sprintf(q, b.DbUser)
	_, err := b.DB.Exec(q)
	if err != nil {
		log.Fatalf("Database reset did not work: %s.", err)
	}
	log.Print("Creating tables.")
	b.migrate()
}

func (b builderio) migrate() {
	b.GormDB.AutoMigrate(
		&db.Item{},
		&db.Page{},
		&db.Part{},
		&db.NameString{},
		&db.PageNameString{},
	)
	db.TruncateBHL(b.DB)
}

func (b builderio) resetDirs() error {
	err := gnsys.CleanDir(b.DownloadDir)
	if err != nil {
		return err
	}
	err = db.ResetKeyVal(b.KeyValDir)
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
	exists, _ := gnsys.FileExists(b.DownloadFile)
	if exists {
		return os.Remove(b.DownloadFile)
	}
	return nil
}
