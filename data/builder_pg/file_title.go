package builder_pg

import (
	"bufio"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Title struct {
	ID        int
	Name      string
	YearStart sql.NullInt32
	YearEnd   sql.NullInt32
	Language  string
	DOI       string
}

const (
	idF        = 0
	nameF      = 4
	yearStartF = 7
	yearEndF   = 8
	langF      = 9
)

func (b BuilderPG) prepareTitle(doiMap map[int]string) (map[int]*Title, error) {
	log.Println("Processing title.txt")
	res := make(map[int]*Title)
	path := filepath.Join(b.Config.DownloadDir, "title.txt")
	f, err := os.Open(path)
	if err != nil {
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
			return res, err
		}

		t := &Title{
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
		return res, err
	}
	return res, nil
}
