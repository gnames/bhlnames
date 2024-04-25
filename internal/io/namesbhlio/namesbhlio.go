package namesbhlio

import (
	"context"
	"encoding/csv"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/gnames/bhlnames/internal/ent/namebhl"
	"github.com/gnames/bhlnames/internal/io/dbio"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

const (
	occurBatchSize = 100_000
	nameBatchSize  = 50_000
)

type namesbhlio struct {
	cfg    config.Config
	db     *pgxpool.Pool
	gormDB *gorm.DB
}

func New(cfg config.Config, db *pgxpool.Pool, gormdb *gorm.DB) namebhl.NameBHL {
	res := namesbhlio{cfg: cfg, db: db, gormDB: gormdb}
	return res
}

// ImportOccurrences transfers occurrences data from bhlindex's
// occurrences.csv dump file to the database.
func (n namesbhlio) ImportOccurrences(blf *bloom.BloomFilter) error {
	slog.Info("Importing names' occurrences.")
	slog.Info("Truncating data from name_occurrences table.")
	err := dbio.Truncate(n.db, []string{"name_occurrences"})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	chOccur := make(chan []model.NameOccurrence)

	g.Go(func() error {
		return n.saveOcurrences(ctx, chOccur, blf)
	})

	err = n.loadOccurrences(chOccur)
	if err != nil {
		return err
	}
	close(chOccur)

	err = g.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (n namesbhlio) loadOccurrences(chIn chan<- []model.NameOccurrence) error {
	path := filepath.Join(n.cfg.ExtractDir, "occurrences.csv")
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)

	_, err = r.Read()
	if err != nil {
		slog.Error("Could not read header of occurrences.csv.", "error", err)
		return err
	}

	chunk := make([][]string, occurBatchSize)
	var count int

	var row []string
	var occurs []model.NameOccurrence
	for {
		row, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("Could not read a row from occurrences.csv.", "error", err)
			return err
		}
		if count == occurBatchSize {
			occurs, err = convertToOccurs(chunk)
			if err != nil {
				return err
			}
			chIn <- occurs
			chunk = make([][]string, occurBatchSize)
			count = 0
		}
		chunk[count] = row
		count++
	}
	occurs, err = convertToOccurs(chunk[0:count])
	if err != nil {
		return err
	}
	chIn <- occurs
	return nil
}

const (
	occNameIDF               = 0
	occPageIDF               = 1
	occItemIDF               = 2
	occDetectedNameF         = 3
	occDetectedNameVerbatimF = 4
	occOddsLog10F            = 5
	occStartF                = 6
	occEndF                  = 7
	occEndsNextPageF         = 8
	occAnnotation            = 9
)

func convertToOccurs(data [][]string) ([]model.NameOccurrence, error) {
	var err error
	res := make([]model.NameOccurrence, 0, len(data))

	for _, v := range data {
		var start, end, pageID int
		start, err = strconv.Atoi(v[occStartF])
		if err != nil {
			slog.Error("Could not convert start to int.", "start", v[occStartF])
			return nil, err
		}
		end, err = strconv.Atoi(v[occEndF])
		if err != nil {
			slog.Error("Could not convert end to int.", "end", v[occEndF])
			return nil, err
		}
		pageID, err = strconv.Atoi(v[occPageIDF])
		if err != nil {
			slog.Error("Could not convert page_id to int.", "page_id", v[occPageIDF])
			return nil, err
		}
		var odds float64
		if v[occOddsLog10F] != "" {
			odds, err = strconv.ParseFloat(v[occOddsLog10F], 64)
			if err != nil {
				slog.Error("Could not convert odds to float.", "odds", v[occOddsLog10F])
				return nil, err
			}
		}
		oc := model.NameOccurrence{
			NameStringID: v[occNameIDF],
			PageID:       uint(pageID),
			OffsetStart:  uint(start),
			OffsetEnd:    uint(end),
			OddsLog10:    odds,
			AnnotNomen:   v[occAnnotation],
		}
		res = append(res, oc)
	}

	return res, nil
}

// ImportNames takes batches of verified names from a bhlindex dump file
// and saves them into database.
func (n namesbhlio) ImportNames() (*bloom.BloomFilter, error) {
	slog.Info("Ingesting names resolved to Catalogue of Life.")

	// blf is a Bloom filter that will be used to check if a name string.
	// is already in the database. If it is, we will not add it again.
	// The filter is created with an estimated number of elements and
	// a false positive rate of 0.1%.
	blf := bloom.NewWithEstimates(25_000_000, 0.001)

	slog.Info("Truncating data from names_strings table.")
	err := dbio.Truncate(n.db, []string{"name_strings"})
	if err != nil {
		slog.Error("Could not truncate name_strings table.", "error", err)
		return blf, err
	}

	chIn := make(chan [][]string)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err = n.loadNames(ctx, chIn)
		close(chIn)
		if err != nil {
			slog.Error("Could not load name-strings.", "error", err)
			return err
		}
		return nil
	})

	g.Go(func() error {
		err = n.saveNames(ctx, chIn, blf)
		if err != nil {
			for range chIn {
			}
		}
		return err
	})

	if err = g.Wait(); err != nil {
		return blf, err
	}

	return blf, nil
}

func (n namesbhlio) loadNames(
	ctx context.Context,
	chIn chan<- [][]string,
) error {
	path := filepath.Join(n.cfg.ExtractDir, "names.csv")
	f, err := os.Open(path)
	if err != nil {
		slog.Error("Could not open names.csv.", "error", err)
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)

	_, err = r.Read()
	if err != nil {
		slog.Error("Could not read header of names.csv.", "error", err)
		return err
	}

	chunk := make([][]string, nameBatchSize)
	var count int

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("Could not read a row from names.csv.", "error", err)
			return err
		}
		if count == nameBatchSize {
			chIn <- chunk
			chunk = make([][]string, nameBatchSize)
			count = 0
		}
		chunk[count] = row
		count++

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	chIn <- chunk[0:count]
	return nil
}
