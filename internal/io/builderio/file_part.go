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
	partIDF             = 0
	partItemIDF         = 1
	partContributorF    = 2
	partSeqOrderF       = 3
	partSegTypeF        = 4
	partTitleF          = 5
	partContainerTitleF = 6
	partPublDetF        = 7
	partVolF            = 8
	partSeriesF         = 9
	partIssueF          = 10
	partDateF           = 11
	partPageRangeF      = 12
	partPageIDF         = 13
	partLangF           = 14
)

type partDate struct {
	year    sql.NullInt32
	yearEnd sql.NullInt32
	month   sql.NullInt32
	day     sql.NullInt32
}

type partPages struct {
	first  sql.NullInt32
	last   sql.NullInt32
	length sql.NullInt32
}

var dateRe = regexp.MustCompile(`\b([\d]{4})\b\s*(-\s*([\d]{1,4})\b(-([\d]{1,2}))?)?`)
var pagesRe = regexp.MustCompile(`\b([\d]+)\b\s*((,|-|--|–)\s*\b([\d]+)\b)?`)

func (b builderio) importPart(doiMap map[int]string) error {
	slog.Info("Preparing part.txt data for db.")
	//keeps unique IDs of the parts
	pMap := make(map[int]struct{})
	var res []*model.Part
	path := filepath.Join(b.cfg.DownloadDir, "part.txt")
	f, err := os.Open(path)
	if err != nil {
		slog.Error("Cannot open part.txt.", "path", path, "error", err)
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
		part := model.Part{}
		l := scanner.Text()
		fields := strings.Split(l, "\t")

		id, err := strconv.Atoi(fields[partIDF])
		if err != nil {
			slog.Error("Cannot convert part id to int.", "id", fields[partIDF])
			return err
		}
		if _, ok := pMap[id]; ok {
			continue
		} else {
			pMap[id] = struct{}{}
		}
		part.ID = uint(id)

		pageID, err := strconv.Atoi(fields[partPageIDF])
		if err == nil {
			part.PageID = sql.NullInt32{Int32: int32(pageID), Valid: true}
		}

		itemID, err := strconv.Atoi(fields[partItemIDF])
		if err == nil {
			part.ItemID = sql.NullInt32{Int32: int32(itemID), Valid: true}
		}

		seqOrder, err := strconv.Atoi(fields[partSeqOrderF])
		if err == nil {
			part.SequenceOrder = sql.NullInt32{Int32: int32(seqOrder), Valid: true}
		}
		part.DOI = doiMap[id]
		part.ContributorName = fields[partContributorF]
		part.SegmentType = fields[partSegTypeF]
		part.Title = fields[partTitleF]
		part.ContainerTitle = fields[partContainerTitleF]
		part.PublicationDetails = fields[partPublDetF]
		part.Volume = fields[partVolF]
		part.Series = fields[partSeriesF]
		part.Issue = fields[partIssueF]
		part.Language = fields[partLangF]

		part.Date = fields[partDateF]
		d := parseDate(part.Date)
		part.Year = d.year
		part.YearEnd = d.yearEnd
		part.Month = d.month
		part.Day = d.day

		pages := parsePages(fields[partPageRangeF])
		part.PageNumStart = pages.first
		part.PageNumEnd = pages.last
		part.Length = pages.length

		res = append(res, &part)

		if err != nil {
			slog.Error("Error converting part data.", "error", err)
		}
	}
	return b.importParts(res)
}

func (b builderio) importParts(parts []*model.Part) error {
	// Part has PageID for the first page of the part.
	// It has page rage, with start and end page numbers provided by publisher.
	// PageID and page range might be empty, or badly formed.
	// We try our best to give PageID and Length, which gives us a range of
	// pages for the part.

	// Additional complication is that PageIDs do not always go in sequence,
	// pages have sequence number, which allows to provide correct PageIDs.
	var err error
	slog.Info(
		"Importing records to parts table",
		"records-num", humanize.Comma(int64(len(parts))),
	)
	columns := []string{"id", "page_id", "item_id", "length", "doi",
		"contributor_name", "sequence_order", "segment_type", "title",
		"container_title", "publication_details", "volume", "series",
		"issue", "date", "year", "year_end", "month", "day", "page_num_start",
		"page_num_end", "language"}
	rows := make([][]any, len(parts))

	for i, v := range parts {
		row := []any{v.ID, v.PageID, v.ItemID, v.Length, v.DOI,
			v.ContributorName, v.SequenceOrder, v.SegmentType, v.Title,
			v.ContainerTitle, v.PublicationDetails, v.Volume, v.Series,
			v.Issue, v.Date, v.Year, v.YearEnd, v.Month, v.Day,
			v.PageNumStart, v.PageNumEnd, v.Language}
		rows[i] = row
	}

	_, err = dbio.InsertRows(b.db, "parts", columns, rows)
	if err != nil {
		slog.Error("Error inserting rows to parts table.", "error", err)
		return err
	}

	return nil
}

func parsePages(pgs string) partPages {
	res := partPages{}
	match := pagesRe.FindStringSubmatch(pgs)
	if match == nil {
		return res
	}
	num, _ := strconv.Atoi(match[1])
	res.first = sql.NullInt32{Int32: int32(num), Valid: true}
	num, err := strconv.Atoi(match[4])
	if err != nil {
		return res
	}
	last := sql.NullInt32{Int32: int32(num), Valid: true}
	size := last.Int32 - res.first.Int32
	if size > 0 {
		res.last = last
		res.length = sql.NullInt32{Int32: size, Valid: true}
	}
	return res
}

func parseDate(date string) partDate {
	res := partDate{}
	match := dateRe.FindStringSubmatch(date)
	if match == nil {
		return res
	}
	num, _ := strconv.Atoi(match[1])
	res.year = sql.NullInt32{Int32: int32(num), Valid: true}
	num, err := strconv.Atoi(match[3])
	if err != nil {
		return res
	}
	if num <= 12 {
		res.month = sql.NullInt32{Int32: int32(num), Valid: true}
	} else if num > 999 {
		res.yearEnd = sql.NullInt32{Int32: int32(num), Valid: true}
	}
	num, err = strconv.Atoi(match[5])
	if err != nil || num > 31 {
		return res
	}
	res.day = sql.NullInt32{Int32: int32(num), Valid: true}
	return res
}
