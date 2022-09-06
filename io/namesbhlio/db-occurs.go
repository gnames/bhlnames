package namesbhlio

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/io/db"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

const OccurBatchSize = 50_000

func (n namesbhlio) saveOcurrences(chIn <-chan []db.NameOccurrence) error {
	columns := []string{"page_id", "name_string_id", "offset_start",
		"offset_end", "odds_log10", "nomen_annot"}
	transaction, err := n.db.Begin()
	if err != nil {
		err = fmt.Errorf("saveOccurrences: %w", err)
		log.Fatal().Err(err)
		return err
	}
	stmt, err := transaction.Prepare(pq.CopyIn("name_occurrences", columns...))
	if err != nil {
		log.Fatal().Err(err)
		return fmt.Errorf("saveOccurrences: %w", err)
	}

	var count int
	var missing int
	for ocs := range chIn {
		count += len(ocs)
		missing += OccurBatchSize - len(ocs)
		fmt.Printf("\r%s", strings.Repeat(" ", 47))
		fmt.Printf("\rImported %s occurrences to db", humanize.Comma(int64(count)))
		for i := range ocs {
			_, err = stmt.Exec(ocs[i].PageID, ocs[i].NameStringID,
				ocs[i].OffsetStart, ocs[i].OffsetEnd, ocs[i].OddsLog10,
				ocs[i].NomenAnnot)
			if err != nil {
				return fmt.Errorf("saveOccurrences: %w", err)
			}
		}
	}

	err = stmt.Close()
	if err != nil {
		return fmt.Errorf("saveOccurrences: %w", err)
	}

	err = transaction.Commit()
	if err != nil {
		return fmt.Errorf("saveOccurrences: %w", err)
	}
	fmt.Println()
	log.Info().Msgf(
		"Imported %s name occurrences, %s occurrences ignored (no page reference).",
		humanize.Comma(int64(count)),
		humanize.Comma(int64(missing)),
	)
	return nil
}
