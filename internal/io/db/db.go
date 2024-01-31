package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gnames/bhlnames/pkg/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func opts(cfg config.Config) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DbHost, cfg.DbUser, cfg.DbPass, cfg.DbDatabase)
}

func NewDbGorm(cnf config.Config) (*gorm.DB, error) {
	db, err := gorm.Open("postgres", opts(cnf))
	if err != nil {
		err = fmt.Errorf("db.NewDbGorm: %#w", err)
		slog.Error("Cannot connect to DB", "error", err)
		return nil, err
	}
	return db, nil
}

func NewDB(cnf config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", opts(cnf))
	if err != nil {
		err = fmt.Errorf("db.NewDB: %#w", err)
		slog.Error("Cannot connect to DB", "error", err)
		return nil, err
	}
	return db, nil
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

func RunQuery(d *sql.DB, q string) (*sql.Rows, error) {
	rows, err := d.Query(q)
	if err != nil {
		err = fmt.Errorf("db.RunQuery: %#w", err)
		slog.Error("Cannot run query", "query", q, "error", err)
		return nil, err
	}
	return rows, nil
}

// QuoteString makes a string value compatible with SQL synthax by wrapping it
// in quotes and escaping internal quotes.
func QuoteString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
