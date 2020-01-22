package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type DbOpts struct {
	Host string
	User string
	Pass string
	Name string
}

func (do DbOpts) NewDbGorm() *gorm.DB {
	db, err := gorm.Open("postgres", do.opts())
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func (do DbOpts) opts() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		do.Host, do.User, do.Pass, do.Name)
}

func (do DbOpts) NewDb() *sql.DB {
	db, err := sql.Open("postgres", do.opts())
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func TruncateBHL(d *gorm.DB) {
	tables := []string{"items", "pages", "parts"}
	for _, v := range tables {
		q := fmt.Sprintf("TRUNCATE TABLE %s", v)
		d.Exec(q)
	}
}

func TruncateNames(d *sql.DB) error {
	tables := []string{"name_strings", "page_name_strings"}
	for _, v := range tables {
		q := fmt.Sprintf("TRUNCATE TABLE %s", v)
		_, err := d.Exec(q)
		if err != nil {
			return err
		}
	}
	return nil
}

func TruncateOccur(d *sql.DB) error {
	tables := []string{"page_name_strings"}
	for _, v := range tables {
		q := fmt.Sprintf("TRUNCATE TABLE %s", v)
		_, err := d.Exec(q)
		if err != nil {
			return err
		}
	}
	return nil
}

func RunQuery(d *sql.DB, q string) *sql.Rows {
	rows, err := d.Query(q)
	if err != nil {
		log.Fatal(err)
	}
	return rows
}
