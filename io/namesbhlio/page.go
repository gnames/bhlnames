package namesbhlio

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/io/db"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type item struct {
	ID    string
	Pages []string
}

// PageFilesToIDs maps filename of a page to the BHL's page ID.
// Mapping of a page filename to BHLs database page id requires three steps.
// 1. During import of page data we create key-value store that keeps field
// `SequenceOrder` called FileNum: `44|mobot123`. Each key has page ID as
// value.
// 2. Then we sort page files by their names for each item and lookup
// the key-value store for page position in array using key from 1.
// 3. We use filename number as a key, and database page ID as a value and
// save them to the same key-value store.
//
// Now the key-value store has mixed keys that look like `44|mobot123` and
// `41*mobot123`, pointing to page IDs as values.
func (n namesbhlio) PageFilesToIDs() (err error) {
	log.Info().Msg("Finding IDs for Page files.")
	kvP := db.InitKeyVal(n.cfg.PageDir)
	defer kvP.Close()
	chIn := make(chan item)
	eg := errgroup.Group{}

	eg.Go(func() error {
		return n.savePageMaps(kvP, chIn)
	})

	err = n.loadPages(chIn)
	if err != nil {
		return fmt.Errorf("PageFilesToIDs: %w", err)
	}
	close(chIn)

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("PageFilesToIDs: %w", err)
	}
	return nil
}

func (n namesbhlio) loadPages(chIn chan<- item) error {
	path := filepath.Join(n.cfg.DownloadDir, "pages.csv")
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

	i := item{}
	var count int
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal().Err(err).Msg("loadPages:")
			return err
		}

		itm, pg := row[0], row[1]
		if i.ID == "" {
			i.ID = itm
		} else if i.ID != itm {
			count++
			chIn <- i
			i = item{ID: itm}
			if count%1000 == 0 {
				fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 47))
				fmt.Fprintf(os.Stderr, "\rProcessed page ID mappings for %s items", humanize.Comma(int64(count)))
			}
		}
		i.Pages = append(i.Pages, pg)
	}
	count++
	chIn <- i
	fmt.Fprintln(os.Stderr)
	log.Info().Msgf("Processed all page IDs for %d items.", count)
	return nil
}

func (n namesbhlio) savePageMaps(kv *badger.DB, chIn <-chan item) error {
	res := make(map[string]int)
	var count int
	for itm := range chIn {
		count++
		for i := range itm.Pages {
			key := strconv.Itoa(i+1) + "|" + itm.ID
			pageID := db.GetValue(kv, key)
			if pageID > 0 {
				fileID := itm.Pages[i] + "*" + itm.ID
				// Assign key to page number taken from file.
				// Assign value to the databsae page id.
				// It allows to do a lookup of a database page IDs knowing only
				// the filename of a page.
				res[fileID] = pageID
			}
		}
		if count == 500 {
			err := savePageIDs(kv, res)
			if err != nil {
				err = fmt.Errorf("savePageMaps: %w", err)
				log.Fatal().Err(err)
				return err
			}
			res = make(map[string]int)
			count = 0
		}
	}

	err := savePageIDs(kv, res)
	if err != nil {
		err = fmt.Errorf("savePageMaps: %w", err)
		log.Fatal().Err(err)
		return err
	}
	return nil
}

func savePageIDs(kv *badger.DB, ids map[string]int) error {
	var err error
	kvTxn := kv.NewTransaction(true)

	for k, v := range ids {
		key := []byte(k)
		val := []byte(strconv.Itoa(int(v)))
		if err = kvTxn.Set(key, val); err == badger.ErrTxnTooBig {
			err = kvTxn.Commit()
			if err != nil {
				return fmt.Errorf("savePageIDs: %w", err)
			}

			kvTxn = kv.NewTransaction(true)
			err = kvTxn.Set(key, val)
			if err != nil {
				return fmt.Errorf("savePageIDs: %w", err)
			}
		}
	}

	err = kvTxn.Commit()
	if err != nil {
		return fmt.Errorf("savePageIDs: %w", err)
	}
	return nil
}
