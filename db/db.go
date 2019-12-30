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

func Truncate(d *gorm.DB) error {
	var tbl string
	rows, err := d.Raw("select tablename from pg_tables where schemaname='public'").Rows()
	if err != nil {
		return err
	}
	for rows.Next() {
		rows.Scan(&tbl)
		q := fmt.Sprintf("TRUNCATE TABLE %s", tbl)
		d.Exec(q)
	}
	return rows.Close()
}
