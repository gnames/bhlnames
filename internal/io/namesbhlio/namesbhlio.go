package namesbhlio

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/gnames/bhlnames/internal/ent/namebhl"
	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/jinzhu/gorm"
	"golang.org/x/sync/errgroup"
)

const (
	occurBatchSize = 100_000
	nameBatchSize  = 50_000
)

type namesbhlio struct {
	cfg    config.Config
	client *http.Client
	db     *sql.DB
	gormDB *gorm.DB
}

func New(cfg config.Config, db *sql.DB, gormdb *gorm.DB) namebhl.NameBHL {
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 10 * time.Second,
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	res := namesbhlio{cfg: cfg, client: client, db: db, gormDB: gormdb}
	return res
}

// ImportOccurrences transfers occurrences data from bhlindex's
// occurrences.csv dump file to the database.
func (n namesbhlio) ImportOccurrences(blf *bloom.BloomFilter) error {
	slog.Info("Importing names' occurrences.")
	slog.Info("Truncating data from name_occurrences table.")
	err := db.Truncate(n.db, []string{"name_occurrences"})
	if err != nil {
		return err
	}

	g := errgroup.Group{}

	chOccur := make(chan []db.NameOccurrence)

	g.Go(func() error {
		return n.saveOcurrences(chOccur, blf)
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

func (n namesbhlio) loadOccurrences(chIn chan<- []db.NameOccurrence) error {
	path := filepath.Join(n.cfg.DownloadDir, "occurrences.csv")
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)

	_, err = r.Read()
	if err != nil {
		return err
	}

	chunk := make([][]string, occurBatchSize)
	var count int

	var row []string
	var occurs []db.NameOccurrence
	for {
		row, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if count == occurBatchSize {
			occurs, err = convertToOccurs(chunk)
			if err != nil {
				return fmt.Errorf("convertToOccurs: %w", err)
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
		return fmt.Errorf("convertToOccurs: %w", err)
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

func convertToOccurs(data [][]string) ([]db.NameOccurrence, error) {
	var err error
	res := make([]db.NameOccurrence, 0, len(data))

	for _, v := range data {
		var start, end, pageID int
		start, err = strconv.Atoi(v[occStartF])
		if err != nil {
			return nil, err
		}
		end, err = strconv.Atoi(v[occEndF])
		if err != nil {
			return nil, err
		}
		pageID, err = strconv.Atoi(v[occPageIDF])
		if err != nil {
			return nil, err
		}
		var odds float64
		if v[occOddsLog10F] != "" {
			odds, err = strconv.ParseFloat(v[occOddsLog10F], 64)
			if err != nil {
				return nil, err
			}
		}
		oc := db.NameOccurrence{
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
	slog.Info("Truncating data from names_strings table.")

	blf := bloom.NewWithEstimates(25_000_000, 0.001)

	err := db.Truncate(n.db, []string{"name_strings"})
	if err != nil {
		return blf, err
	}

	chIn := make(chan [][]string)
	eg := &errgroup.Group{}
	egDB := &errgroup.Group{}

	eg.Go(func() error {
		err = n.loadNames(chIn)
		if err != nil {
			err = fmt.Errorf("loadNames: %w", err)
		}
		return err
	})

	egDB.Go(func() error {
		err = n.saveNames(chIn, blf)
		if err != nil {
			err = fmt.Errorf("saveNames: %w", err)
		}
		return err
	})

	if err = eg.Wait(); err != nil {
		return blf, err
	}
	close(chIn)

	if err = egDB.Wait(); err != nil {
		return blf, err
	}
	return blf, nil
}

func (n namesbhlio) loadNames(chIn chan<- [][]string) error {
	path := filepath.Join(n.cfg.DownloadDir, "names.csv")
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)

	_, err = r.Read()
	if err != nil {
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
			return err
		}
		if count == nameBatchSize {
			chIn <- chunk
			chunk = make([][]string, nameBatchSize)
			count = 0
		}
		chunk[count] = row
		count++
	}
	chIn <- chunk[0:count]
	return nil
}
