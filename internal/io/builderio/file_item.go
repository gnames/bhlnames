package builderio

import (
	"bufio"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/gnames/bhlnames/internal/io/dbio"
)

const (
	itemIDF      = 0
	itemTitleIDF = 1
	itemBarCodeF = 3
	itemVolF     = 6
	itemYearsF   = 12
)

var yrRe = regexp.MustCompile(`\b[c]?([\d]{4})\b\s*([,/-]\s*([\d]{4})\b)?`)

// importItem reads item.txt file and imports data to the items table.
// It takes a map of titles as input, and uses it to add title data to the item.
// the key of the map is title id, the value contains a title data.
func (b builderio) importItem(titles map[int]*model.Title) error {
	slog.Info("Preparing item.txt data for db.")
	iMap := make(map[int]struct{})
	var res []*model.Item
	path := filepath.Join(b.cfg.DownloadDir, "item.txt")
	f, err := os.Open(path)
	if err != nil {
		slog.Error("Cannot open item.txt.", "path", path, "error", err)
		return err
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

		var id, titleID int

		id, err = strconv.Atoi(fields[itemIDF])
		if err != nil {
			slog.Error("Cannot convert item id to int.", "id", fields[itemIDF])
			return err
		}
		if _, ok := iMap[id]; ok {
			continue
		} else {
			iMap[id] = struct{}{}
		}
		titleID, err = strconv.Atoi(fields[itemTitleIDF])
		if err != nil {
			slog.Error("Cannot convert title id to int.", "id", fields[itemTitleIDF])
			return err
		}

		barCode := fields[itemBarCodeF]
		vol := fields[itemVolF]
		yearStart, yearEnd := itemYears(fields[itemYearsF])
		t := titles[titleID]
		if t == nil {
			t = &model.Title{}
		}
		item := model.Item{ID: uint(id), TitleID: uint(titleID), TitleDOI: t.DOI,
			BarCode: barCode, Vol: vol, YearStart: yearStart, YearEnd: yearEnd,
			TitleName: t.Name, TitleYearStart: t.YearStart, TitleYearEnd: t.YearEnd,
			TitleLang: t.Language}
		res = append(res, &item)
	}

	if err = scanner.Err(); err != nil {
		slog.Error("Error reading item.txt.", "error", err)
		return err
	}

	err = b.importItems(res)
	if err != nil {
		return err
	}
	return nil
}

func (b builderio) importItems(items []*model.Item) error {
	slog.Info("Importing records to items table", "records-num", humanize.Comma(int64(len(items))))
	columns := []string{"id", "bar_code", "vol", "year_start", "year_end",
		"title_id", "title_doi", "title_name", "title_year_start", "title_year_end",
		"title_lang"}

	rows := make([][]any, len(items))
	for i, v := range items {
		row := []any{v.ID, v.BarCode, v.Vol, v.YearStart, v.YearEnd,
			v.TitleID, v.TitleDOI, v.TitleName, v.TitleYearStart, v.TitleYearEnd,
			v.TitleLang}
		rows[i] = row
	}
	_, err := dbio.InsertRows(b.db, "items", columns, rows)
	if err != nil {
		slog.Error("Cannot insert items.", "error", err)
		return err
	}

	return nil
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
