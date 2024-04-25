package builderio

import (
	"bufio"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/gnames/bhlnames/internal/io/dbio"
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

func (b builderio) importPage() error {
	var err error
	var id, itemID, fileNum, pageNum int
	slog.Info("Importing page.txt data to db.")

	total := 0
	pMap := make(map[int]struct{})
	res := make([]*model.Page, 0, BatchSize)
	path := filepath.Join(b.cfg.ExtractDir, "page.txt")
	f, err := os.Open(path)
	if err != nil {
		slog.Error("Cannot open page.txt.", "path", path, "error", err)
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

		id, err = strconv.Atoi(fields[pageIDF])
		if err != nil {
			slog.Error("Cannot convert page id to int.", "id", fields[pageIDF])
			return err
		}

		if _, ok := pMap[id]; ok {
			continue
		} else {
			pMap[id] = struct{}{}
		}
		count++
		page := &model.Page{ID: uint(id)}

		itemID, err = strconv.Atoi(fields[pageItemIDF])
		if err != nil {
			slog.Error("Cannot convert item id to int.", "id", fields[pageItemIDF])
			return err
		}
		page.ItemID = uint(itemID)

		fileNum, err = strconv.Atoi(fields[pageFileNumF])
		if err != nil {
			slog.Error(
				"Cannot convert file number to int.",
				"file number", fields[pageFileNumF],
			)
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
			pages := make([]*model.Page, len(res))
			copy(pages, res)
			err = b.processPages(pages, total)
			if err != nil {
				return err
			}
			res = make([]*model.Page, 0, BatchSize)
		}
	}
	total += len(res)
	err = b.processPages(res, total)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 35))
	slog.Info(
		"Imported page.txt data to db.",
		"records-num", humanize.Comma(int64(total)),
	)
	return nil
}

func (b builderio) processPages(
	pages []*model.Page,
	total int,
) error {
	var err error
	columns := []string{"id", "item_id", "sequence_order", "page_num"}
	rows := make([][]any, len(pages))

	for i, v := range pages {
		row := []any{v.ID, v.ItemID, v.SequenceOrder, v.PageNum}
		rows[i] = row
	}

	_, err = dbio.InsertRows(b.db, "pages", columns, rows)
	if err != nil {
		slog.Error("Error inserting rows to pages table.", "error", err)
		return err
	}

	fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 35))
	fmt.Fprintf(os.Stderr, "\rImported %s pages to db", humanize.Comma(int64(total)))
	return nil
}
