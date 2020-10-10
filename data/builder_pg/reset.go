package builder_pg

import (
	"fmt"
	"log"
	"os"

	"github.com/gnames/bhlnames/db"
	"github.com/gnames/gnames/lib/sys"
)

func (b BuilderPG) resetDB() {
	q := `
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO %s;
COMMENT ON SCHEMA public IS 'standard public schema'`
	q = fmt.Sprintf(q, b.Config.DB.User)
	_, err := b.DB.Exec(q)
	if err != nil {
		log.Fatalf("Database reset did not work: %s.", err)
	}
}

func (b BuilderPG) migrate() {
	d := b.GormDB
	d.AutoMigrate(
		&db.Item{},
		&db.Page{},
		&db.Part{},
		&db.NameString{},
		&db.PageNameString{},
	)
	db.TruncateBHL(d)
}

func (b BuilderPG) resetDirs() error {
	fs := b.Config.FileSystem
	err := sys.CleanDir(fs.DownloadDir)
	if err != nil {
		return err
	}
	err = db.ResetKeyVal(fs.KeyValDir)
	if err != nil {
		return err
	}
	err = db.ResetKeyVal(fs.PartDir)
	if err != nil {
		return err
	}
	if sys.FileExists(fs.DownloadFile) {
		return os.Remove(fs.DownloadFile)
	}
	return nil
}
