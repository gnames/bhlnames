package bhl

import (
	"bufio"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gnames/bhlnames/db"
	"github.com/lib/pq"
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
	year    sql.NullInt64
	yearEnd sql.NullInt64
	month   sql.NullInt64
	day     sql.NullInt64
}

type partPages struct {
	first  sql.NullInt64
	last   sql.NullInt64
	length sql.NullInt64
}

var dateRe = regexp.MustCompile(`\b([\d]{4})\b\s*(-\s*([\d]{1,4})\b(-([\d]{1,2}))?)?`)
var pagesRe = regexp.MustCompile(`\b([\d]+)\b\s*((,|-|--|–)\s*\b([\d]+)\b)?`)

func (md MetaData) uploadPart(doiMap map[int]string) error {
	log.Println("Preparing part.txt data for db.")
	pMap := make(map[int]struct{})
	var res []*db.Part
	path := filepath.Join(md.DownloadDir, "part.txt")
	f, err := os.Open(path)
	if err != nil {
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
		part := db.Part{}
		l := scanner.Text()
		fields := strings.Split(l, "\t")

		id, err := strconv.Atoi(fields[partIDF])
		if err != nil {
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
			part.PageID = sql.NullInt64{Int64: int64(pageID), Valid: true}
		}

		itemID, err := strconv.Atoi(fields[partItemIDF])
		if err == nil {
			part.ItemID = sql.NullInt64{Int64: int64(itemID), Valid: true}
		}

		seqOrder, err := strconv.Atoi(fields[partSeqOrderF])
		if err == nil {
			part.SequenceOrder = sql.NullInt64{Int64: int64(seqOrder), Valid: true}
		}

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
	}
	return md.uploadParts(res)
}

func (md MetaData) uploadParts(items []*db.Part) error {
	log.Printf("Uploading %d records to parts table.", len(items))
	columns := []string{"id", "page_id", "item_id", "length", "doi",
		"contributor_name", "sequence_order", "segment_type", "title",
		"container_title", "publication_details", "volume", "series",
		"issue", "date", "year", "year_end", "month", "day", "page_num_start",
		"page_num_end", "language"}
	transaction, err := md.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := transaction.Prepare(pq.CopyIn("parts", columns...))
	if err != nil {
		return err
	}

	for _, v := range items {
		_, err = stmt.Exec(v.ID, v.PageID, v.ItemID, v.Length, v.DOI,
			v.ContributorName, v.SequenceOrder, v.SegmentType, v.Title,
			v.ContainerTitle, v.PublicationDetails, v.Volume, v.Series,
			v.Issue, v.Date, v.Year, v.YearEnd, v.Month, v.Day,
			v.PageNumStart, v.PageNumEnd, v.Language)
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
	return transaction.Commit()
}

func parsePages(pgs string) partPages {
	res := partPages{}
	match := pagesRe.FindStringSubmatch(pgs)
	if match == nil {
		return res
	}
	num, _ := strconv.Atoi(match[1])
	res.first = sql.NullInt64{Int64: int64(num), Valid: true}
	num, err := strconv.Atoi(match[4])
	if err != nil {
		return res
	}
	last := sql.NullInt64{Int64: int64(num), Valid: true}
	size := last.Int64 - res.first.Int64
	if size > 0 {
		res.last = last
		res.length = sql.NullInt64{Int64: size, Valid: true}
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
	res.year = sql.NullInt64{Int64: int64(num), Valid: true}
	num, err := strconv.Atoi(match[3])
	if err != nil {
		return res
	}
	if num <= 12 {
		res.month = sql.NullInt64{Int64: int64(num), Valid: true}
	} else if num > 999 {
		res.yearEnd = sql.NullInt64{Int64: int64(num), Valid: true}
	}
	num, err = strconv.Atoi(match[5])
	if err != nil || num > 31 {
		return res
	}
	res.day = sql.NullInt64{Int64: int64(num), Valid: true}
	return res
}
