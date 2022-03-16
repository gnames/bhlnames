package builderio

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/io/db"
	"github.com/lib/pq"
)

const (
	pageIDF      = 0
	pageItemIDF  = 1
	pageFileNumF = 2
	pageNumberF  = 7
)

const BatchSize = 100_000

func (b builderio) importPage(itemMap map[uint]string) error {
	log.Println("Uploading page.txt data for db.")
	err := db.ResetKeyVal(b.Config.PageDir)
	if err != nil {
		return err
	}
	kv := db.InitKeyVal(b.Config.PageDir)
	total := 0
	pMap := make(map[int]struct{})
	res := make([]*db.Page, 0, BatchSize)
	path := filepath.Join(b.Config.DownloadDir, "page.txt")
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

		id, err := strconv.Atoi(fields[pageIDF])
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

		itemID, err := strconv.Atoi(fields[pageItemIDF])
		if err != nil {
			return err
		}
		page.ItemID = uint(itemID)

		fileNum, err := strconv.Atoi(fields[pageFileNumF])
		if err != nil {
			return err
		}
		page.FileNum = uint(fileNum)

		pageNum, err := strconv.Atoi(fields[pageNumberF])
		if err == nil {
			page.PageNum = sql.NullInt64{Int64: int64(pageNum), Valid: true}
		}
		res = append(res, page)

		if count >= BatchSize {
			count = 0
			total += len(res)
			pages := make([]*db.Page, len(res))
			copy(pages, res)
			err := b.processPages(kv, itemMap, pages, total)
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
	columns := []string{"id", "item_id", "file_num", "page_num"}
	transaction, err := b.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := transaction.Prepare(pq.CopyIn("pages", columns...))
	if err != nil {
		return err
	}
	kvTxn := kv.NewTransaction(true)

	for _, v := range pages {
		key := fmt.Sprintf("%d|%s", v.FileNum, itemMap[v.ItemID])
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

		_, err = stmt.Exec(v.ID, v.ItemID, v.FileNum, v.PageNum)
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
