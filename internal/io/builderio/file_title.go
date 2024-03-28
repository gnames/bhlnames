package builderio

import (
	"bufio"
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/jackc/pgx/v5"
)

const (
	idF        = 0
	nameF      = 4
	yearStartF = 7
	yearEndF   = 8
	langF      = 9
)

// prepareTitle reads title.txt file and prepares a map of titles.
// It takes a map of DOI ids as input, and uses it to add DOI to the title.
func (b builderio) prepareTitle(doiMap map[int]string) (map[int]*model.Title, error) {
	slog.Info("Processing title.txt.")
	res := make(map[int]*model.Title)
	path := filepath.Join(b.cfg.DownloadDir, "title.txt")
	f, err := os.Open(path)
	if err != nil {
		slog.Error("Cannot open title.txt.", "error", err)
		return res, err
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
		id, err := strconv.Atoi(fields[idF])
		if err != nil {
			slog.Error("Cannot convert title id to int.", "id", fields[idF])
			return res, err
		}

		t := &model.Title{
			ID:       id,
			Name:     fields[nameF],
			Language: fields[langF],
		}
		ys, err := strconv.Atoi(fields[yearStartF])
		if err == nil {
			t.YearStart = sql.NullInt32{Int32: int32(ys), Valid: true}
		} else {
			t.YearStart = sql.NullInt32{Valid: false}
		}

		ye, err := strconv.Atoi(fields[yearEndF])
		if err == nil {
			t.YearEnd = sql.NullInt32{Int32: int32(ye), Valid: true}
		} else {
			t.YearEnd = sql.NullInt32{Valid: false}
		}

		t.DOI = doiMap[t.ID]
		res[t.ID] = t
	}
	if err := scanner.Err(); err != nil {
		slog.Error("Error reading title.txt.", "error", err)
		return res, err
	}

	return res, nil
}

func (b *builderio) dbTitlesMap() (map[int]*model.Title, error) {
	res := make(map[int]*model.Title)
	var err error
	var rows pgx.Rows
	rows, err = b.db.Query(
		context.Background(),
		`
SELECT
	title_id, title_name, title_year_start, title_year_end,
	title_lang, title_doi
FROM items `,
	)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		t := model.Title{}
		err = rows.Scan(
			&t.ID, &t.Name, &t.YearStart, &t.YearEnd, &t.Language, &t.DOI,
		)
		if err != nil {
			return res, err
		}
		res[t.ID] = &t
	}

	return res, nil
}
