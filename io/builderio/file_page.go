package builderio

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/io/db"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

const (
	// pageIDF is an automatically assigned page ID from BHL database.
	// It is sequence that starts from 1 and autoincrmented by 1.
	pageIDF = 0
	// pageItemIDF is an automatically assigned item ID from BHL database.
	pageItemIDF = 1
	// pageFileNumF is a sequential number of page in an item. Sadly it
	// does not correspond to number given in a page filename.
	pageFileNumF = 2
	// pageNumberF is a 'real' page number given to a page by a publisher of
	// an item. We need to connect this number to a number from a page
	// filename. We can do it if we know which pageID corresponds to which
	// number extracted from a page filename.
	pageNumberF = 7
)

const BatchSize = 100_000

func (b builderio) importPage(itemMap map[uint]string) error {
	var err error
	var id, itemID, fileNum, pageNum int
	log.Info().Msg("Importing page.txt data to db.")
	err = db.ResetKeyVal(b.PageDir)
	if err != nil {
		return err
	}
	kv := db.InitKeyVal(b.PageDir)
	total := 0
	pMap := make(map[int]struct{})
	res := make([]*db.Page, 0, BatchSize)
	path := filepath.Join(b.DownloadDir, "page.txt")
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	defer kv.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	header := true
	for scanner.Scan() {
		if header {
			header = false
			continue
		}
		l := scanner.Text()
		fields := strings.Split(l, "\t")

		id, err = strconv.Atoi(fields[pageIDF])
		if err != nil {
			return err
		}

		if _, ok := pMap[id]; ok {
			continue
		} else {
			pMap[id] = struct{}{}
		}
		count++
		page := &db.Page{ID: uint(id)}

		itemID, err = strconv.Atoi(fields[pageItemIDF])
		if err != nil {
			return err
		}
		page.ItemID = uint(itemID)

		fileNum, err = strconv.Atoi(fields[pageFileNumF])
		if err != nil {
			return err
		}
		page.SequenceOrder = uint(fileNum)

		pageNum, err = strconv.Atoi(fields[pageNumberF])
		if err == nil {
			page.PageNum = sql.NullInt64{Int64: int64(pageNum), Valid: true}
		}
		res = append(res, page)

		if count >= BatchSize {
			count = 0
			total += len(res)
			pages := make([]*db.Page, len(res))
			copy(pages, res)
			err = b.processPages(kv, itemMap, pages, total)
			if err != nil {
				return err
			}
			res = res[:0]

		}
	}
	total += len(res)
	err = b.processPages(kv, itemMap, res, total)
	fmt.Println()
	return err
}

func (b builderio) processPages(
	kv *badger.DB,
	itemMap map[uint]string,
	pages []*db.Page,
	total int,
) error {
	columns := []string{"id", "item_id", "sequence_order", "page_num"}
	transaction, err := b.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := transaction.Prepare(pq.CopyIn("pages", columns...))
	if err != nil {
		return err
	}
	kvTxn := kv.NewTransaction(true)

	// we create a dataset with the following schema:
	// key: "page sequence number|itemBarcode". For example "13|tropicosabc-3"
	// value pageID (from the database)
	//
	// If we have an item from BHL filesystem with pages, we can calculate
	// more or less accurate which page corresponds to which pageID by
	// sorting pages according to filenames, and figuring out which number
	// taken from the file-name corresponds to which pageID.
	//
	// This method is not perfect, but is good for most of the items. A better
	// way would be to have the number exctracted from file in the page.txt
	// dump.
	for _, v := range pages {
		key := fmt.Sprintf("%d|%s", v.SequenceOrder, itemMap[v.ItemID])
		val := strconv.Itoa(int(v.ID))
		if err = kvTxn.Set([]byte(key), []byte(val)); err == badger.ErrTxnTooBig {
			err = kvTxn.Commit()
			if err != nil {
				return err
			}

			kvTxn = kv.NewTransaction(true)
			err = kvTxn.Set([]byte(key), []byte(val))
			if err != nil {
				return err
			}
		}

		_, err = stmt.Exec(v.ID, v.ItemID, v.SequenceOrder, v.PageNum)
		if err != nil {
			return err
		}
	}

	err = kvTxn.Commit()
	if err != nil {
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rImported %s pages to db", humanize.Comma(int64(total)))
	return transaction.Commit()
}
