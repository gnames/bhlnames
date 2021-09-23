package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/gnames/bhlnames/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func opts(cfg config.Config) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DbHost, cfg.DbUser, cfg.DbPass, cfg.DbName)
}

func NewDbGorm(cnf config.Config) *gorm.DB {
	db, err := gorm.Open("postgres", opts(cnf))
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func NewDB(cnf config.Config) *sql.DB {
	db, err := sql.Open("postgres", opts(cnf))
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func TruncateBHL(d *sql.DB) {
	tables := []string{"items", "pages", "parts"}
	for _, v := range tables {
		q := fmt.Sprintf("TRUNCATE TABLE %s", v)
		_, err := d.Exec(q)
		if err != nil {
			log.Fatalf("Cannot truncate table '%s': %s", v, err)
		}
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

// QuoteString makes a string value compatible with SQL synthax by wrapping it
// in quotes and escaping internal quotes.
func QuoteString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
