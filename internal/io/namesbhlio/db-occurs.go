package namesbhlio

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/gnames/bhlnames/internal/io/dbio"
)

func (n namesbhlio) saveOcurrences(
	ctx context.Context,
	chIn <-chan []model.NameOccurrence,
	blf *bloom.BloomFilter,
) error {
	var i int
	columns := []string{"page_id", "name_string_id", "offset_start",
		"offset_end", "odds_log10", "annot_nomen"}
	var count, missing int
	for ocs := range chIn {
		rows := make([][]any, 0, len(ocs))
		for i = range ocs {
			// do not save occurrences for which there is no verified name saved
			if !blf.Test([]byte(ocs[i].NameStringID)) {
				missing++
				continue
			}

			row := []any{ocs[i].PageID, ocs[i].NameStringID,
				ocs[i].OffsetStart, ocs[i].OffsetEnd, ocs[i].OddsLog10,
				ocs[i].AnnotNomen}

			rows = append(rows, row)
		}

		_, err := dbio.InsertRows(n.db, "name_occurrences", columns, rows)
		if err != nil {
			slog.Error("Cannot insert rows to name_occurrences table", "error", err)
			for range chIn {
			}
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			count += len(rows)
			fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 47))
			fmt.Fprintf(os.Stderr, "\rImported %s occurrences.", humanize.Comma(int64(count)))
		}
	}
	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 47))
	slog.Info("Imported name occurrences.",
		"records-num", humanize.Comma(int64(count)),
		"ignored-num", humanize.Comma(int64(missing)),
	)

	return nil
}
