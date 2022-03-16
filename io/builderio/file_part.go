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

	badger "github.com/dgraph-io/badger/v2"
	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/io/db"
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
	log.Println("Preparing part.txt data for db.")
	//keeps unique IDs of the parts
	pMap := make(map[int]struct{})
	var res []*db.Part
	path := filepath.Join(b.Config.DownloadDir, "part.txt")
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	err = db.ResetKeyVal(b.Config.PartDir)
	if err != nil {
		return err
	}
	kv := db.InitKeyVal(b.Config.PartDir)
	defer f.Close()
	defer kv.Close()
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
	}
	return b.uploadParts(kv, res)
}

func (b builderio) uploadParts(kv *badger.DB, items []*db.Part) error {
	log.Printf("Uploading %s records to parts table.", humanize.Comma(int64(len(items))))
	columns := []string{"id", "page_id", "item_id", "length", "doi",
		"contributor_name", "sequence_order", "segment_type", "title",
		"container_title", "publication_details", "volume", "series",
		"issue", "date", "year", "year_end", "month", "day", "page_num_start",
		"page_num_end", "language"}
	transaction, err := b.DB.Begin()
	if err != nil {
		return err
	}
	kvTxn := kv.NewTransaction(true)

	stmt, err := transaction.Prepare(pq.CopyIn("parts", columns...))
	if err != nil {
		return err
	}

	for _, v := range items {
		if v.PageID.Valid {
			length := int(v.Length.Int32)
			for i := 0; i <= length; i++ {
				key := strconv.Itoa(int(v.PageID.Int32) + i)
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
			}
		}
		_, err = stmt.Exec(v.ID, v.PageID, v.ItemID, v.Length, v.DOI,
			v.ContributorName, v.SequenceOrder, v.SegmentType, v.Title,
			v.ContainerTitle, v.PublicationDetails, v.Volume, v.Series,
			v.Issue, v.Date, v.Year, v.YearEnd, v.Month, v.Day,
			v.PageNumStart, v.PageNumEnd, v.Language)
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
	return transaction.Commit()
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
