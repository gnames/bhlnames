package namesbhlio

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/lib/pq"
)

func (n namesbhlio) saveOcurrences(
	chIn <-chan []db.NameOccurrence,
	blf *bloom.BloomFilter,
) error {
	columns := []string{"page_id", "name_string_id", "offset_start",
		"offset_end", "odds_log10", "annot_nomen"}
	var count, missing int
	for ocs := range chIn {
		transaction, err := n.db.Begin()
		if err != nil {
			return err
		}
		stmt, err := transaction.Prepare(pq.CopyIn("name_occurrences", columns...))
		if err != nil {
			return err
		}

		for i := range ocs {
			// do not save occurrences for which there is no verified name saved
			if !blf.Test([]byte(ocs[i].NameStringID)) {
				continue
			}

			_, err = stmt.Exec(ocs[i].PageID, ocs[i].NameStringID,
				ocs[i].OffsetStart, ocs[i].OffsetEnd, ocs[i].OddsLog10,
				ocs[i].AnnotNomen)
			if err != nil {
				return err
			}
		}

		count += len(ocs)
		missing += occurBatchSize - len(ocs)
		fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 47))
		fmt.Fprintf(os.Stderr, "\rImported %s occurrences.", humanize.Comma(int64(count)))

		err = stmt.Close()
		if err != nil {
			return err
		}

		err = transaction.Commit()
		if err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr)
	str := fmt.Sprintf(
		"Imported %s name occurrences, %s occurrences ignored (no page reference).",
		humanize.Comma(int64(count)),
		humanize.Comma(int64(missing)),
	)
	slog.Info(str)

	return nil
}
