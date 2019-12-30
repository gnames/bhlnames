package bhl

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gnames/bhlnames/db"
	"github.com/lib/pq"
)

const (
	pageIDF      = 0
	pageItemIDF  = 1
	pageFileNumF = 2
	pageNumberF  = 7
)

const BatchSize = 500000

func (md MetaData) uploadPage() error {
	log.Println("Uploading page.txt data for db.")
	total := 0
	pMap := make(map[int]struct{})
	res := make([]*db.Page, 0, BatchSize)
	path := filepath.Join(md.DownloadDir, "page.txt")
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
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
			err := md.uploadPages(pages, total)
			if err != nil {
				return err
			}
			res = res[:0]

		}
	}
	total += len(res)
	err = md.uploadPages(res, total)
	fmt.Println()
	return err
}

func (md MetaData) uploadPages(items []*db.Page, total int) error {
	columns := []string{"id", "item_id", "file_num", "page_num"}
	transaction, err := md.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := transaction.Prepare(pq.CopyIn("pages", columns...))
	if err != nil {
		return err
	}

	for _, v := range items {
		_, err = stmt.Exec(v.ID, v.ItemID, v.FileNum, v.PageNum)
		if err != nil {
			return err
		}
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
	fmt.Printf("\rUploaded %d pages to db", total)
	return transaction.Commit()
}
