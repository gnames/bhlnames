package reffinderio

import (
	"database/sql"
	"fmt"

	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/io/db"
	"github.com/rs/zerolog/log"
)

type preReference struct {
	item *row
	part *db.Part
}

type row struct {
	itemID             int
	titleID            int
	pageID             int
	pageNum            int
	titleDOI           string
	titleYearStart     int
	titleYearEnd       int
	yearStart          int
	yearEnd            int
	volume             string
	titleName          string
	mainTaxon          string
	mainKingdom        string
	mainKingdomPercent int
	namesTotal         int
	nameID             string
	name               string
	annotation         string
	matchedCanonical   string
	matchType          string
	editDistance       int
}

func (l reffinderio) nameOnlyOccurrences(nameRefs *namerefs.NameRefs) []*row {
	return l.occurrences(nameRefs.Canonical, "matched_canonical")
}

func (l reffinderio) taxonOccurrences(nameRefs *namerefs.NameRefs) []*row {
	return l.occurrences(nameRefs.CurrentCanonical, "current_canonical")
}

func (l reffinderio) occurrences(name string, field string) []*row {
	switch field {
	case "matched_canonical", "current_canonical":
	default:
		log.Warn().Msgf("Unregistered field: %s", field)
		return nil
	}
	var res []*row
	var itemID, titleID, pageID int
	var yearStart, yearEnd, titleYearStart, titleYearEnd, pageNum,
		kingdomPercent, pathsTotal, editDistance sql.NullInt32
	var nameID string
	var titleName, context, majorKingdom, nameString, matchedCanonical,
		matchType, vol, titleDOI, annot sql.NullString
	qs := `SELECT
  itm.id, itm.title_id, pns.page_id, pg.page_num, pns.annot_nomen,
  itm.title_year_start, itm.title_year_end, itm.year_start, itm.year_end,
  itm.title_name, itm.vol, itm.title_doi, ist.main_taxon, ist.main_kingdom,
  ist.main_kingdom_percent, ist.names_total, ns.id, ns.name, ns.matched_canonical,
  ns.match_type, ns.edit_distance
	FROM name_strings ns
			JOIN name_occurrences pns ON ns.id = pns.name_string_id
			JOIN pages pg ON pg.id = pns.page_id
			JOIN items itm ON itm.id = pg.item_id
      JOIN item_stats ist ON itm.id = ist.id
	WHERE ns.%s = $1
	ORDER BY title_year_start`
	q := fmt.Sprintf(qs, field)

	rows, err := l.DB.Query(q, name)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot find occurrences.")
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&itemID, &titleID, &pageID, &pageNum, &annot,
			&titleYearStart, &titleYearEnd, &yearStart, &yearEnd, &titleName, &vol,
			&titleDOI, &context, &majorKingdom, &kingdomPercent, &pathsTotal,
			&nameID, &nameString, &matchedCanonical, &matchType, &editDistance)
		if err != nil {
			err = fmt.Errorf("reffinderio.occurrences: %w", err)
			log.Fatal().Err(err)
		}
		res = append(res, &row{
			itemID:             itemID,
			titleID:            titleID,
			pageID:             pageID,
			pageNum:            int(pageNum.Int32),
			titleDOI:           titleDOI.String,
			titleYearStart:     int(titleYearStart.Int32),
			titleYearEnd:       int(titleYearEnd.Int32),
			yearStart:          int(yearStart.Int32),
			yearEnd:            int(yearEnd.Int32),
			titleName:          titleName.String,
			volume:             vol.String,
			mainTaxon:          context.String,
			mainKingdom:        majorKingdom.String,
			mainKingdomPercent: int(kingdomPercent.Int32),
			namesTotal:         int(pathsTotal.Int32),
			nameID:             nameID,
			name:               nameString.String,
			annotation:         annot.String,
			matchedCanonical:   matchedCanonical.String,
			matchType:          matchType.String,
			editDistance:       int(editDistance.Int32),
		})
	}
	return res
}

func (l reffinderio) currentCanonical(canonical string) (string, error) {
	var currentCan sql.NullString
	q := `SELECT current_canonical
          FROM name_strings
        WHERE matched_canonical = $1
        LIMIT 1`
	row := l.DB.QueryRow(q, canonical)
	err := row.Scan(&currentCan)
	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return currentCan.String, nil
}
