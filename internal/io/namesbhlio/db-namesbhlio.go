package namesbhlio

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/internal/io/dbio"
)

const (
	NameIDF              = 0
	DetectedNameF        = 1
	CardinalityF         = 2
	OccurrencesNumberF   = 3
	OddsLog10F           = 4
	MatchTypeF           = 5
	MatchSortOrderF      = 6
	EditDistanceF        = 7
	StemEditDistanceF    = 8
	MatchedCanonicalF    = 9
	MatchedFullNameF     = 10
	MatchedCardinalityF  = 11
	CurrentCanonicalF    = 12
	CurrentFullNameF     = 13
	CurrentCardinalityF  = 14
	ClassificationF      = 15
	ClassificationRanksF = 16
	ClassificationIDsF   = 17
	RecordIDF            = 18
	DataSourceIDF        = 19
	DataSourceF          = 20
	DataSourcesNumberF   = 21
	CurationF            = 22
	ErrorF               = 23
)

func (n namesbhlio) saveNames(
	ctx context.Context,
	ch <-chan [][]string,
	blf *bloom.BloomFilter,
) error {
	var err error
	total := 0
	columns := []string{"id", "name", "record_id", "match_type",
		"match_sort_order", "edit_distance", "stem_edit_distance", "matched_name",
		"matched_canonical", "current_name", "current_canonical", "classification",
		"classification_ranks", "classification_ids", "data_source_id",
		"data_source_title", "data_sources_number", "curation", "occurences",
		"odds_log10", "error"}

	for names := range ch {
		total += len(names)

		var eDist, stemDist, dsID, matchSort, dsNum, occurs int
		var odds float64

		rows := make([][]any, 0, len(names))
		for _, v := range names {
			eDist, err = strconv.Atoi(v[EditDistanceF])
			if err == nil {
				stemDist, err = strconv.Atoi(v[StemEditDistanceF])
			}
			if err == nil {
				dsID, err = strconv.Atoi(v[DataSourceIDF])
			}
			if err == nil {
				matchSort, err = strconv.Atoi(v[MatchSortOrderF])
			}
			if err == nil {
				dsNum, err = strconv.Atoi(v[DataSourcesNumberF])
			}
			if err == nil {
				occurs, err = strconv.Atoi(v[OccurrencesNumberF])
			}
			if err != nil {
				slog.Error("Cannot convert string to int", "error", err)
				for range ch {
				}
				return err
			}

			// for now only take Catalogue of Life names.
			if dsID != 1 {
				continue
			}

			blf.Add([]byte(v[NameIDF]))

			odds, err = strconv.ParseFloat(v[OddsLog10F], 64)
			if err != nil {
				odds = 0
			}

			row := []any{v[NameIDF], v[DetectedNameF], v[RecordIDF],
				v[MatchTypeF], matchSort, eDist, stemDist, v[MatchedFullNameF],
				v[MatchedCanonicalF], v[CurrentFullNameF], v[CurrentCanonicalF],
				v[ClassificationF], v[ClassificationRanksF], v[ClassificationIDsF],
				dsID, v[DataSourceF], dsNum, true, occurs, odds, v[ErrorF]}
			rows = append(rows, row)
		}
		_, err = dbio.InsertRows(n.db, "name_strings", columns, rows)
		if err != nil {
			slog.Error("Cannot insert rows to name_strings table", "error", err)
			for range ch {
			}
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 47))
			fmt.Fprintf(os.Stderr, "\rImported %s names to db", humanize.Comma(int64(total)))
		}
	}
	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 47))
	slog.Info("Imported names to db", "records-num", humanize.Comma(int64(total)))
	return nil
}
