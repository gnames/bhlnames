package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gnames/bhlnames/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func opts(cnf config.DB) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		cnf.Host, cnf.User, cnf.Pass, cnf.Name)
}

func NewDbGorm(cnf config.DB) *gorm.DB {
	db, err := gorm.Open("postgres", opts(cnf))
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func NewDb(cnf config.DB) *sql.DB {
	log.Printf("Connecting to Postgres DB at %s", cnf.Host)
	db, err := sql.Open("postgres", opts(cnf))
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
