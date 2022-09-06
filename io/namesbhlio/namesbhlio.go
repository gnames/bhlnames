package namesbhlio

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/namebhl"
	"github.com/gnames/bhlnames/io/db"
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
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
func (n namesbhlio) ImportOccurrences() error {
	log.Info().Msg("Importing names' occurrences.")
	log.Info().Msg("Truncating data from name_occurrences table.")
	err := db.Truncate(n.db, []string{"name_occurrences"})
	if err != nil {
		return fmt.Errorf("ImportOccurrences: %w", err)
	}

	g := errgroup.Group{}

	chOccur := make(chan []db.NameOccurrence)

	g.Go(func() error {
		return n.saveOcurrences(chOccur)
	})

	err = n.loadOccurrences(chOccur)
	if err != nil {
		return fmt.Errorf("ImportOccurrences: %w", err)
	}
	close(chOccur)

	err = g.Wait()
	if err != nil {
		return fmt.Errorf("ImportOccurrences: %w", err)
	}
	return nil
}

func (n namesbhlio) loadOccurrences(chIn chan<- []db.NameOccurrence) error {
	kv := db.InitKeyVal(n.cfg.PageDir)
	defer kv.Close()

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

	chunk := make([][]string, OccurBatchSize)
	var count int

	var row []string
	var occurs []db.NameOccurrence
	for {
		row, err = r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal().Err(err).Msg("loadOccurrences:")
			return err
		}
		if count == OccurBatchSize {
			occurs, err = convertToOccurs(kv, chunk)
			if err != nil {
				log.Fatal().Err(err).Msg("loadOccurrences:")
				return err
			}
			chIn <- occurs
			chunk = make([][]string, OccurBatchSize)
			count = 0
		}
		chunk[count] = row
		count++
	}
	occurs, err = convertToOccurs(kv, chunk[0:count])
	if err != nil {
		log.Fatal().Err(err).Msg("loadOccurrences:")
		return err
	}
	chIn <- occurs
	return nil
}

const (
	occItemBarcodeF          = 0
	occPageBarcodeNumF       = 1
	occNameIDF               = 2
	occDetectedNameF         = 3
	occDetectedNameVerbatimF = 4
	occOddsLog10F            = 5
	occStartF                = 6
	occEndF                  = 7
	occEndsNextPageF         = 8
	occAnnotation            = 9
)

func convertToOccurs(kv *badger.DB, data [][]string) ([]db.NameOccurrence, error) {
	var err error
	resPrelim := make([]db.NameOccurrence, len(data))
	res := make([]db.NameOccurrence, 0, len(data))
	keys := make([]string, len(data))

	for i, v := range data {
		var start, end int
		start, err = strconv.Atoi(v[occStartF])
		if err != nil {
			err = fmt.Errorf("convertToOccurs: %w", err)
			return nil, err
		}
		end, err = strconv.Atoi(v[occEndF])
		if err != nil {
			err = fmt.Errorf("convertToOccurs: %w", err)
			return nil, err
		}
		var odds float64
		if v[occOddsLog10F] != "" {
			odds, err = strconv.ParseFloat(v[occOddsLog10F], 64)
			if err != nil {
				err = fmt.Errorf("convertToOccurs: %w", err)
				return nil, err
			}
		}

		oc := db.NameOccurrence{
			NameStringID: v[occNameIDF],
			OffsetStart:  uint(start),
			OffsetEnd:    uint(end),
			OddsLog10:    odds,
			NomenAnnot:   v[occAnnotation],
		}
		keys[i] = v[occPageBarcodeNumF] + "*" + v[occItemBarcodeF]
		resPrelim[i] = oc
	}

	ids, err := db.GetValues(kv, keys)
	if err != nil {
		err = fmt.Errorf("convertToOccurs: %w", err)
		return nil, err
	}

	for i := range keys {
		if bs := ids[keys[i]]; len(bs) > 0 {
			id, err := strconv.Atoi(string(bs))
			if err != nil {
				err = fmt.Errorf("convertToOccurs: %w", err)
				return nil, err
			}
			resPrelim[i].PageID = uint(id)
			res = append(res, resPrelim[i])
		}
	}
	return res, nil
}

// ImportNames takes batches of verified names from a bhlindex dump file
// and saves them into database.
func (n namesbhlio) ImportNames() error {
	log.Info().Msg("Ingesting names resolved to Catalogue of Life.")
	log.Info().Msg("Truncating data from names_strings table.")

	err := db.Truncate(n.db, []string{"name_strings"})
	if err != nil {
		return fmt.Errorf("importNames: %w", err)
	}

	chIn := make(chan [][]string)
	eg := &errgroup.Group{}
	egDB := &errgroup.Group{}

	eg.Go(func() error {
		return n.loadNames(chIn)
	})

	egDB.Go(func() error {
		return n.saveNames(chIn)
	})

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("importNames: %w", err)
	}
	close(chIn)

	if err = egDB.Wait(); err != nil {
		return fmt.Errorf("importNames: %w", err)
	}
	return err
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

	chunkSize := 50_000
	chunk := make([][]string, chunkSize)
	var count int

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal().Err(err).Msg("loadNames:")
			return err
		}
		if count == chunkSize {
			chIn <- chunk
			chunk = make([][]string, chunkSize)
			count = 0
		}
		chunk[count] = row
		count++
	}
	chIn <- chunk[0:count]
	return nil
}
