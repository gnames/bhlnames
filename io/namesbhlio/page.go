package namesbhlio

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlindex/ent/page"
	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/gnfmt"
	"github.com/rs/zerolog/log"
)

// PageFilesToIDs maps name of the file of a page to the BHL's page ID.
func (n namesbhlio) PageFilesToIDs() (err error) {
	log.Info().Msg("Finding ID for Page Files")
	kvP := db.InitKeyVal(n.cfg.PageDir)
	defer kvP.Close()

	offsetID := 1
	limit := 100
	var ps map[string][]page.Page
	for {
		ps, err = n.getPages(offsetID, limit)
		if len(ps) == 0 {
			break
		}
		err = processPageIDs(kvP, ps)
		if err != nil {
			return fmt.Errorf("PageFilesToIDs: %w", err)
		}
		offsetID += limit
		fmt.Printf("\r%s", strings.Repeat(" ", 35))
		fmt.Printf("\rGot page IDs for %s items", humanize.Comma(int64(offsetID-1)))
	}
	return nil
}

func processPageIDs(kv *badger.DB, ps map[string][]page.Page) error {
	res := make(map[string]int)
	for k, v := range ps {
		for i := range v {
			key := strconv.Itoa(i+1) + "|" + k
			pageID := db.GetValue(kv, key)
			if pageID > 0 {
				res[v[i].ID] = pageID
			}
		}
	}

	err := savePageIDs(kv, res)
	if err != nil {
		return fmt.Errorf("processPageIDs: %w", err)
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

func (n namesbhlio) getPages(offsetID, limit int) (map[string][]page.Page, error) {
	var pages []page.Page
	namesURL := fmt.Sprintf("%spages?limit=%d&offset_id=%d",
		n.cfg.BHLIndexURL, limit, offsetID)

	respBytes, err := n.getREST(namesURL)
	if err != nil {
		return nil, fmt.Errorf("getPages: %w", err)
	}

	err = gnfmt.GNjson{}.Decode(respBytes, &pages)
	if err != nil {
		return nil, fmt.Errorf("getPages: %w", err)
	}
	return itemPages(pages), nil
}

func itemPages(ps []page.Page) map[string][]page.Page {
	res := make(map[string][]page.Page)
	for i := range ps {
		l := len(ps[i].ID)
		barCode := ps[i].ID[0 : l-5]
		res[barCode] = append(res[barCode], ps[i])
	}
	return res
}
