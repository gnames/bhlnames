package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gnames/bhlnames/pkg/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rs/zerolog/log"
)

func opts(cfg config.Config) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DbHost, cfg.DbUser, cfg.DbPass, cfg.DbDatabase)
}

func NewDbGorm(cnf config.Config) *gorm.DB {
	db, err := gorm.Open("postgres", opts(cnf))
	if err != nil {
		err = fmt.Errorf("db.NewDbGorm: %#w", err)
		log.Fatal().Err(err).Msg("NewDbGorm")
	}
	return db
}

func NewDB(cnf config.Config) *sql.DB {
	db, err := sql.Open("postgres", opts(cnf))
	if err != nil {
		err = fmt.Errorf("db.NewDB: %#w", err)
		log.Fatal().Err(err).Msg("NewDB")
	}
	return db
}

func Truncate(d *sql.DB, tables []string) error {
	for _, v := range tables {
		q := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY", v)
		_, err := d.Exec(q)
		if err != nil {
			return fmt.Errorf("cannot truncate '%s': %w", v, err)
		}
	}
	return nil
}

func RunQuery(d *sql.DB, q string) *sql.Rows {
	rows, err := d.Query(q)
	if err != nil {
		err = fmt.Errorf("db.RunQuery: %#w", err)
		log.Fatal().Err(err).Msg("RunQuery")
	}
	return rows
}

// QuoteString makes a string value compatible with SQL synthax by wrapping it
// in quotes and escaping internal quotes.
func QuoteString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
