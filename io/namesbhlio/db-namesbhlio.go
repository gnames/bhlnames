package namesbhlio

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

const (
	NameIDF             = 0
	DetectedNameF       = 1
	CardinalityF        = 2
	OccurrencesNumberF  = 3
	OddsLog10F          = 4
	MatchTypeF          = 5
	EditDistanceF       = 6
	StemEditDistanceF   = 7
	MatchedCanonicalF   = 8
	MatchedFullNameF    = 9
	MatchedCardinalityF = 10
	CurrentCanonicalF   = 11
	CurrentFullNameF    = 12
	CurrentCardinalityF = 13
	ClassificationF     = 14
	RecordIDF           = 15
	DataSourceIDF       = 16
	DataSourceF         = 17
	DataSourcesNumberF  = 18
	CurationF           = 19
	ErrorF              = 20
)

func (n namesbhlio) saveNames(ch <-chan [][]string) error {
	total := 0

	for names := range ch {
		total += len(names)
		columns := []string{"id", "name", "record_id", "match_type",
			"edit_distance", "stem_edit_distance", "matched_name", "matched_canonical",
			"current_name", "current_canonical", "classification", "data_source_id",
			"data_source_title", "data_sources_number", "curation", "occurences",
			"odds_log10", "error"}
		transaction, err := n.db.Begin()
		if err != nil {
			return err
		}
		stmt, err := transaction.Prepare(pq.CopyIn("name_strings", columns...))
		if err != nil {
			return err
		}

		var eDist, stemDist, dsID, dsNum, occurs int
		var odds float64

		for _, v := range names {
			eDist, err = strconv.Atoi(v[EditDistanceF])
			if err != nil {
				log.Fatal().Err(err).Msg("saveNames:")
				return err
			}

			stemDist, err = strconv.Atoi(v[StemEditDistanceF])
			if err != nil {
				log.Fatal().Err(err).Msg("saveNames:")
				return err
			}

			dsID, err = strconv.Atoi(v[DataSourceIDF])
			if err != nil {
				log.Fatal().Err(err).Msg("saveNames:")
				return err
			}

			dsNum, err = strconv.Atoi(v[DataSourcesNumberF])
			if err != nil {
				log.Fatal().Err(err).Msg("saveNames:")
				return err
			}

			occurs, err = strconv.Atoi(v[OccurrencesNumberF])
			if err != nil {
				log.Fatal().Err(err).Msg("saveNames:")
				return err
			}

			odds, err = strconv.ParseFloat(v[OddsLog10F], 64)
			if err != nil {
				odds = 0
			}

			_, err = stmt.Exec(v[NameIDF], v[DetectedNameF], v[RecordIDF],
				v[MatchTypeF], eDist, stemDist,
				v[MatchedFullNameF], v[MatchedCanonicalF], v[CurrentFullNameF],
				v[CurrentCanonicalF], v[ClassificationF], dsID,
				v[DataSourceF], dsNum, true,
				occurs, odds, v[ErrorF])
			if err != nil {
				log.Fatal().Err(err).Msg("saveNames:")
				return err
			}
		}
		err = stmt.Close()
		if err != nil {
			return err
		}
		fmt.Printf("\r%s", strings.Repeat(" ", 47))
		fmt.Printf("\rImported %s names to db", humanize.Comma(int64(total)))
		err = transaction.Commit()
		if err != nil {
			return err
		}
	}
	fmt.Println()
	return nil
}
