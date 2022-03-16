package builderio

import (
	"bufio"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/io/db"
	"github.com/lib/pq"
)

const (
	itemIDF      = 0
	itemTitleIDF = 1
	itemBarCodeF = 3
	itemVolF     = 6
	itemYearsF   = 12
)

var yrRe = regexp.MustCompile(`\b[c]?([\d]{4})\b\s*([,/-]\s*([\d]{4})\b)?`)

func (b builderio) importItem(titles map[int]*title) (map[uint]string, error) {
	log.Println("Preparing item.txt data for db.")
	iMap := make(map[int]struct{})
	var res []*db.Item
	path := filepath.Join(b.Config.DownloadDir, "item.txt")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	header := true
	for scanner.Scan() {
		if header {
			header = false
			continue
		}
		l := scanner.Text()
		fields := strings.Split(l, "\t")

		id, err := strconv.Atoi(fields[itemIDF])
		if err != nil {
			return nil, err
		}
		if _, ok := iMap[id]; ok {
			continue
		} else {
			iMap[id] = struct{}{}
		}
		titleID, err := strconv.Atoi(fields[itemTitleIDF])
		if err != nil {
			return nil, err
		}

		barCode := fields[itemBarCodeF]
		vol := fields[itemVolF]
		yearStart, yearEnd := itemYears(fields[itemYearsF])
		t := titles[titleID]
		if t == nil {
			t = &title{}
		}
		item := db.Item{ID: uint(id), TitleID: uint(titleID), TitleDOI: t.DOI,
			BarCode: barCode, Vol: vol, YearStart: yearStart, YearEnd: yearEnd,
			TitleName: t.Name, TitleYearStart: t.YearStart, TitleYearEnd: t.YearEnd,
			TitleLang: t.Language}
		res = append(res, &item)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	var itemMap map[uint]string
	itemMap, err = b.uploadItems(res)
	if err != nil {
		return nil, err
	}
	return itemMap, nil
}

func (b builderio) uploadItems(items []*db.Item) (map[uint]string, error) {
	log.Printf("Uploading %s records to items table.", humanize.Comma(int64(len(items))))
	columns := []string{"id", "bar_code", "vol", "year_start", "year_end",
		"title_id", "title_doi", "title_name", "title_year_start", "title_year_end",
		"title_lang"}
	transaction, err := b.DB.Begin()
	if err != nil {
		return nil, err
	}

	stmt, err := transaction.Prepare(pq.CopyIn("items", columns...))
	if err != nil {
		return nil, err
	}

	itemMap := make(map[uint]string)
	for _, v := range items {
		itemMap[v.ID] = v.BarCode
		_, err = stmt.Exec(v.ID, v.BarCode, v.Vol, v.YearStart, v.YearEnd,
			v.TitleID, v.TitleDOI, v.TitleName, v.TitleYearStart, v.TitleYearEnd,
			v.TitleLang)
		if err != nil {
			return nil, err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, err
	}

	err = stmt.Close()
	if err != nil {
		return nil, err
	}
	err = transaction.Commit()
	if err != nil {
		return nil, err
	}
	return itemMap, nil
}

func itemYears(years string) (sql.NullInt32, sql.NullInt32) {
	finds := yrRe.FindStringSubmatch(years)
	yrStart := ""
	yrEnd := ""
	if len(finds) > 1 {
		yrStart = finds[1]
	}
	if len(finds) > 3 {
		yrEnd = finds[3]
	}
	yearStart := sql.NullInt32{}
	yearEnd := sql.NullInt32{}
	res, err := strconv.Atoi(yrStart)
	if err == nil {
		yearStart = sql.NullInt32{Int32: int32(res), Valid: true}
	}
	res, err = strconv.Atoi(yrEnd)
	if err == nil {
		yearEnd = sql.NullInt32{Int32: int32(res), Valid: true}
	}
	return yearStart, yearEnd
}
