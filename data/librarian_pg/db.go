package librarian_pg

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gnames/bhlnames/db"
	"github.com/gnames/bhlnames/domain/entity"
)

type preReference struct {
	item *row
	part *db.Part
}

type row struct {
	itemID           int
	titleID          int
	pageID           int
	titleDOI         string
	titleYearStart   int
	titleYearEnd     int
	yearStart        int
	yearEnd          int
	volume           string
	titleName        string
	context          string
	kingdom          string
	kingdomPercent   int
	pathsTotal       int
	nameID           string
	name             string
	annotation       string
	matchedCanonical string
	matchType        string
	editDistance     int
}

func (l LibrarianPG) nameOnlyOccurrences(nameRefs *entity.NameRefs) []*row {
	return l.occurrences(nameRefs.Canonical, "matched_canonical")
}

func (l LibrarianPG) taxonOccurrences(nameRefs *entity.NameRefs) []*row {
	return l.occurrences(nameRefs.CurrentCanonical, "current_canonical")
}

func (l LibrarianPG) occurrences(name string, field string) []*row {
	var res []*row
	var itemID, titleID, pageID int
	var yearStart, yearEnd, titleYearStart, titleYearEnd,
		kingdomPercent, pathsTotal, editDistance sql.NullInt32
	var nameID string
	var titleName, context, majorKingdom, nameString, matchedCanonical,
		matchType, vol, titleDOI, annot sql.NullString
	qs := `SELECT
	itm.id, itm.title_id, pns.page_id, pns.annotation_type, itm.title_year_start,
	itm.title_year_end, itm.year_start, itm.year_end, itm.title_name, itm.vol,
	itm.title_doi, itm.context, itm.major_kingdom, itm.kingdom_percent,
	itm.paths_total, ns.id, ns.name, ns.matched_canonical, ns.match_type,
	ns.edit_distance
	FROM name_strings ns
			JOIN page_name_strings pns ON ns.id = pns.name_string_id
			JOIN pages pg ON pg.id = pns.page_id
			JOIN items itm ON itm.id = pg.item_id
	WHERE ns.%s = '%s'
	ORDER BY title_year_start`
	q := fmt.Sprintf(qs, field, name)

	rows := db.RunQuery(l.DB, q)
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&itemID, &titleID, &pageID, &annot, &titleYearStart,
			&titleYearEnd, &yearStart, &yearEnd, &titleName, &vol, &titleDOI,
			&context, &majorKingdom, &kingdomPercent, &pathsTotal, &nameID,
			&nameString, &matchedCanonical, &matchType, &editDistance)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, &row{
			itemID:           itemID,
			titleID:          titleID,
			pageID:           pageID,
			titleDOI:         titleDOI.String,
			titleYearStart:   int(titleYearStart.Int32),
			titleYearEnd:     int(titleYearEnd.Int32),
			yearStart:        int(yearStart.Int32),
			yearEnd:          int(yearEnd.Int32),
			titleName:        titleName.String,
			volume:           vol.String,
			context:          context.String,
			kingdom:          majorKingdom.String,
			kingdomPercent:   int(kingdomPercent.Int32),
			pathsTotal:       int(pathsTotal.Int32),
			nameID:           nameID,
			name:             nameString.String,
			annotation:       annot.String,
			matchedCanonical: matchedCanonical.String,
			matchType:        matchType.String,
			editDistance:     int(editDistance.Int32),
		})
	}
	return res
}

func (l LibrarianPG) currentCanonical(canonical string) (string, error) {
	var currentCan sql.NullString
	q := `SELECT current_canonical
          FROM name_strings
        WHERE matched_canonical = '%s'
        LIMIT 1`
	q = fmt.Sprintf(q, canonical)
	rows := db.RunQuery(l.DB, q)
	for rows.Next() {
		err := rows.Scan(&currentCan)
		if err != nil {
			return "", err
		}
	}
	return currentCan.String, nil
}
