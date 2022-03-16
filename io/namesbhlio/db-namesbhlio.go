package namesbhlio

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/dustin/go-humanize"
	bhlname "github.com/gnames/bhlindex/ent/name"
	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/uuid5"
	"github.com/lib/pq"
)

func (n namesbhlio) saveNames(ch <-chan []bhlname.VerifiedName) error {
	total := 0

	for names := range ch {
		total += len(names)
		columns := []string{"id", "name", "taxon_id", "match_type",
			"edit_distance", "stem_edit_distance", "matched_name", "matched_canonical",
			"current_name", "current_canonical", "classification", "data_source_id",
			"data_source_title", "data_sources_number", "curation", "occurences",
			"odds", "error"}
		transaction, err := n.db.Begin()
		if err != nil {
			return err
		}
		stmt, err := transaction.Prepare(pq.CopyIn("name_strings", columns...))
		if err != nil {
			return err
		}
		var nameID int
		l := len(names)
		if l > 0 {
			nameID = names[l-1].NameID
		}
		for _, v := range names {
			curation := v.Curation == "Curated"
			id := uuid5.UUID5(v.Name).String()

			_, err = stmt.Exec(id, v.Name, v.RecordID, v.MatchType,
				v.EditDistance, v.StemEditDistance, v.MatchedName,
				v.MatchedCanonical, v.CurrentName, v.CurrentCanonical,
				v.Classification, v.DataSourceID, v.DataSourceTitle,
				v.DataSourcesNumber, curation,
				v.Occurrences, v.OddsLog10, v.Error)
			if err != nil {
				return err
			}
		}
		err = stmt.Close()
		if err != nil {
			return err
		}
		if nameID > 0 {
			fmt.Printf("\r%s", strings.Repeat(" ", 47))
			fmt.Printf("\rImported %s names to db, id: %s",
				humanize.Comma(int64(total)), humanize.Comma(int64(nameID)))
		}
		err = transaction.Commit()
		if err != nil {
			return err
		}
	}
	fmt.Println()
	return nil
}

func (n namesbhlio) saveOcurrences(kv *badger.DB, ch <-chan []bhlname.DetectedName) error {
	var total int
	var missing map[string]struct{}
	missingItems, err := os.Create(
		filepath.Join(n.cfg.InputDir, "missing-items.txt"),
	)
	if err != nil {
		return fmt.Errorf("saveOccurrences: %w", err)
	}
	defer missingItems.Close()

	for occurs := range ch {
		ids, err := getPageIDs(kv, occurs)
		if err != nil {
			return fmt.Errorf("saveOccurrences: %w", err)
		}
		missing, err = n.saveOccursToDB(ids, occurs)
		if err != nil {
			return fmt.Errorf("saveOccurrences: %w", err)
		}
		for k := range missing {
			_, err := missingItems.Write([]byte(k + "\n"))
			if err != nil {
				return fmt.Errorf("saveOccurrences: %w", err)
			}
		}
		total += len(occurs)
		fmt.Printf("\r%s", strings.Repeat(" ", 47))
		fmt.Printf("\rImported %s occurrences to db",
			humanize.Comma(int64(total)))
	}
	return nil
}

func getPageIDs(
	kv *badger.DB,
	occurrs []bhlname.DetectedName,
) (map[string][]byte, error) {
	barCodes := make([]string, len(occurrs))
	for i := range occurrs {
		barCodes[i] = occurrs[i].PageID
	}

	return db.GetValues(kv, barCodes)
}

func (n namesbhlio) saveOccursToDB(
	ids map[string][]byte,
	occurs []bhlname.DetectedName,
) (map[string]struct{}, error) {
	missing := make(map[string]struct{})
	var pageID int
	columns := []string{"page_id", "name_string_id", "offset_start",
		"offset_end", "odds", "annotation", "annotation_type"}
	transaction, err := n.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("saveOccursToDB: %w", err)
	}
	stmt, err := transaction.Prepare(pq.CopyIn("page_name_strings", columns...))
	if err != nil {
		return nil, fmt.Errorf("saveOccursToDB: %w", err)
	}

	for _, v := range occurs {
		id := uuid5.UUID5(v.Name).String()
		if ids[v.PageID] == nil {
			l := len(v.PageID)
			itemBarCode := v.PageID[0 : l-5]
			missing[itemBarCode] = struct{}{}
			continue
		}
		pageID, err = strconv.Atoi(string(ids[v.PageID]))
		if err != nil {
			return nil, fmt.Errorf("saveOccursToDB: %w", err)
		}
		_, err = stmt.Exec(uint(pageID), id, v.OffsetStart, v.OffsetEnd,
			v.OddsLog10, v.AnnotNomen, v.AnnotNomenType)
		if err != nil {
			return nil, fmt.Errorf("saveOccursToDB: %w", err)
		}
	}
	err = stmt.Close()
	if err != nil {
		return nil, fmt.Errorf("saveOccursToDB: %w", err)
	}

	err = transaction.Commit()
	if err != nil {
		return nil, fmt.Errorf("saveOccursToDB: %w", err)
	}

	return missing, nil
}
