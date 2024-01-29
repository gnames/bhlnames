package reffinderio

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/bhlnames/internal/ent/refbhl"
	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/rs/zerolog/log"
)

type preReference struct {
	item *refRow
	part *db.Part
}

type refRow struct {
	itemID             int
	titleID            int
	pageID             int
	pageNum            sql.NullInt64
	titleDOI           string
	titleYearStart     sql.NullInt32
	titleYearEnd       sql.NullInt32
	yearStart          sql.NullInt32
	yearEnd            sql.NullInt32
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

func (rf reffinderio) refByPageID(pageID int) (*refbhl.Reference, error) {
	qs := `SELECT
  itm.id, itm.title_id, pg.id, pg.page_num,
  itm.title_year_start, itm.title_year_end, itm.year_start, itm.year_end,
  itm.title_name, itm.vol, itm.title_doi, ist.main_taxon, ist.main_kingdom,
  ist.main_kingdom_percent, ist.names_total
	FROM pages pg
	  JOIN items itm ON itm.id = pg.item_id
      JOIN item_stats ist ON itm.id = ist.id
	WHERE pg.id = $1
	ORDER BY title_year_start`

	r := rf.db.QueryRow(qs, pageID)

	var rr refRow
	err := r.Scan(&rr.itemID, &rr.titleID, &rr.pageID, &rr.pageNum,
		&rr.titleYearStart, &rr.titleYearEnd, &rr.yearStart,
		&rr.yearEnd, &rr.titleName, &rr.volume, &rr.titleDOI,
		&rr.mainTaxon, &rr.mainKingdom, &rr.mainKingdomPercent,
		&rr.namesTotal,
	)
	if err != nil {
		err = fmt.Errorf("reffinderio.refByPageID: %w", err)
		log.Warn().Err(err).Msg("")
		return nil, err
	}

	preRes := preReference{item: &rr}

	partID := findPart(rf.kvDB, pageID)
	if partID > 0 {

		part, err := rf.partByID(partID)
		if err != nil {
			err = fmt.Errorf("reffinderio.refByPageID: %w", err)
			log.Warn().Err(err).Msg("")
			return nil, err
		}
		preRes.part = part
	}
	res := rf.genReferences([]*preReference{&preRes})
	if len(res) == 0 {
		return nil, errors.New("reffinderio.refByPageID: no references found")
	}
	return &res[0].Reference, nil
}

func (rf reffinderio) partByID(partID int) (*db.Part, error) {
	var res db.Part
	q := `SELECT
  id, title, doi, page_num_start, page_num_end, year
	FROM parts
	WHERE id = $1`
	row := rf.db.QueryRow(q, partID)
	err := row.Scan(&res.ID, &res.Title, &res.DOI, &res.PageNumStart,
		&res.PageNumEnd, &res.Year)
	if err != nil {
		err = fmt.Errorf("reffinderio.partByID: %w", err)
		log.Warn().Err(err).Msg("")
	}
	return &res, nil
}

func (l reffinderio) nameOnlyOccurrences(nameRefs *namerefs.NameRefs) []*refRow {
	return l.occurrences(nameRefs.Canonical, "matched_canonical")
}

func (l reffinderio) taxonOccurrences(nameRefs *namerefs.NameRefs) []*refRow {
	return l.occurrences(nameRefs.CurrentCanonical, "current_canonical")
}

func (l reffinderio) occurrences(name string, field string) []*refRow {
	switch field {
	case "matched_canonical", "current_canonical":
	default:
		log.Warn().Msgf("Unregistered field: %s", field)
		return nil
	}
	var res []*refRow
	var itemID, titleID, pageID int
	var kingdomPercent, pathsTotal, editDistance int
	var yearStart, yearEnd, titleYearStart, titleYearEnd sql.NullInt32
	var pageNum sql.NullInt64
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

	rows, err := l.db.Query(q, name)
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
			log.Fatal().Err(err).Msg("")
		}
		rec := &refRow{
			itemID:             itemID,
			titleID:            titleID,
			pageID:             pageID,
			pageNum:            pageNum,
			titleDOI:           titleDOI.String,
			titleYearStart:     titleYearStart,
			titleYearEnd:       titleYearEnd,
			yearStart:          yearStart,
			yearEnd:            yearEnd,
			titleName:          titleName.String,
			volume:             vol.String,
			mainTaxon:          context.String,
			mainKingdom:        majorKingdom.String,
			mainKingdomPercent: kingdomPercent,
			namesTotal:         pathsTotal,
			nameID:             nameID,
			name:               nameString.String,
			annotation:         annot.String,
			matchedCanonical:   matchedCanonical.String,
			matchType:          matchType.String,
			editDistance:       editDistance,
		}

		res = append(res, rec)
	}
	return res
}

func (l reffinderio) currentCanonical(canonical string) (string, error) {
	var currentCan sql.NullString
	q := `SELECT current_canonical
          FROM name_strings
        WHERE matched_canonical = $1
        LIMIT 1`
	row := l.db.QueryRow(q, canonical)
	err := row.Scan(&currentCan)
	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return currentCan.String, nil
}
