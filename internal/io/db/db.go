package db

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/gnames/bhlnames/pkg/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var dbOnce sync.Once
var gormOnce sync.Once

func opts(cfg config.Config) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DbHost, cfg.DbUser, cfg.DbPass, cfg.DbDatabase)
}

// NewDbGorm creates a new gorm.DB connection to a PostgreSQL database.
func NewDbGorm(cnf config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	gormOnce.Do(func() {
		db, err = gorm.Open("postgres", opts(cnf))
	})
	if err != nil {
		err = fmt.Errorf("db.NewDbGorm: %#w", err)
		slog.Error("Cannot connect to DB", "error", err)
		return nil, err
	}
	return db, nil
}

// NewDB creates a new pgxpool.Pool connection to a PostgreSQL database.
func NewDB(cnf config.Config) (*pgxpool.Pool, error) {
	var db *pgxpool.Pool
	var err error

	pgxCfg, err := pgxpool.ParseConfig(opts(cnf))
	if err != nil {
		slog.Error("Cannot parse pgx config", "error", err)
		return nil, err
	}
	pgxCfg.MaxConns = 15

	dbOnce.Do(func() {
		db, err = pgxpool.NewWithConfig(
			context.Background(),
			pgxCfg,
		)
	})
	if err != nil {
		slog.Error("Cannot connect to database", "error", err)
		return nil, err
	}
	return db, nil
}

func Truncate(ctx context.Context, d *pgxpool.Pool, tables []string) error {
	for _, v := range tables {
		q := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY", v)
		_, err := d.Exec(ctx, q)
		if err != nil {
			return fmt.Errorf("cannot truncate '%s': %w", v, err)
		}
	}
	return nil
}

// func RunQuery(ctx context.Context, d *pgxpool.Pool, q string) (pgx.Rows, error) {
// 	rows, err := d.Query(ctx, q)
// 	if err != nil {
// 		err = fmt.Errorf("db.RunQuer: %w", err)
// 		slog.Error("Cannot run query", "query", q, "error", err)
// 		return nil, err
// 	}
// 	return rows, nil
// }

// QuoteString makes a string value compatible with SQL synthax by wrapping it
// in quotes and escaping internal quotes.
func QuoteString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func InsertRows(
	pool *pgxpool.Pool,
	tbl string,
	columns []string,
	rows [][]any,
) (int64, error) {
	copyCount, err := pool.CopyFrom(
		context.Background(),
		pgx.Identifier{tbl},
		columns,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return 0, err
	}

	return int64(copyCount), nil
}
