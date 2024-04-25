package dbio

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gnames/bhlnames/pkg/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func dburl(cfg config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.DbUser, cfg.DbPass, cfg.DbHost, 5432, cfg.DbDatabase)
}

// NewDB creates a new connections pool to the database.
func NewDB(cfg config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(
		context.Background(),
		dburl(cfg),
	)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// NewGORM creates a new GORM connection to the database.
func NewGORM(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dburl(cfg)), &gorm.Config{})
	if err != nil {
		slog.Error("Cannot setup GORM connection.", "error", err)
		return nil, err
	}
	return db, nil
}

// Truncate removes all data from the tables.
func Truncate(d *pgxpool.Pool, tables []string) error {
	for _, v := range tables {
		q := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY", v)
		_, err := d.Exec(context.Background(), q)
		if err != nil {
			slog.Error("Cannot truncate table.", "table", v, "error", err)
			return err
		}
	}
	return nil
}

// InsertRows inserts a batch of rows into a table.
func InsertRows(
	d *pgxpool.Pool,
	tbl string,
	columns []string,
	rows [][]any,
) (int64, error) {
	copyCount, err := d.CopyFrom(
		context.Background(),
		pgx.Identifier{tbl},
		columns,
		pgx.CopyFromRows(rows),
	)
	return int64(copyCount), err
}
