package reffndio

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/gnames/gnfmt"
	"github.com/jackc/pgx/v5"
)

type preReference struct {
	item *refRec
	part *model.Part
}

type refRec struct {
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

func (rf reffndio) refByPageID(pageID int) (*bhl.Reference, error) {
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

	ctx := context.Background()

	r := rf.db.QueryRow(ctx, qs, pageID)

	var rr refRec
	err := r.Scan(&rr.itemID, &rr.titleID, &rr.pageID, &rr.pageNum,
		&rr.titleYearStart, &rr.titleYearEnd, &rr.yearStart,
		&rr.yearEnd, &rr.titleName, &rr.volume, &rr.titleDOI,
		&rr.mainTaxon, &rr.mainKingdom, &rr.mainKingdomPercent,
		&rr.namesTotal,
	)
	if err != nil {
		err = fmt.Errorf("reffinderio.refByPageID: %w", err)
		slog.Error("Cannot find page", "page_id", pageID, "err", err)
		return nil, err
	}

	preRes := preReference{item: &rr}

	part, err := rf.partByID(pageID)
	if err != nil {
		err = fmt.Errorf("reffinderio.refByPageID: %w", err)
		slog.Error("Cannot find part", "page_id", pageID, "err", err)
		return nil, err
	}
	preRes.part = part
	// }
	res := rf.getReferences([]*preReference{&preRes}, false)
	if len(res) == 0 {
		return nil, errors.New("reffinderio.refByPageID: no references found")
	}
	return &res[0].Reference, nil
}

func (rf reffndio) partByID(pageID int) (*model.Part, error) {
	q := `SELECT
  id, title, doi, page_num_start, page_num_end, year
	FROM parts p
	JOIN page_parts pp
		ON p.id = pp.part_id
	WHERE pp.page_id = $1
	`
	ctx := context.Background()
	rows, err := rf.db.Query(ctx, q, pageID)
	if err != nil {
		slog.Error("Cannot run part query", "error", err)
		return nil, err
	}
	defer rows.Close()

	parts := make([]model.Part, 0)
	for rows.Next() {
		var res model.Part
		err := rows.Scan(&res.ID, &res.Title, &res.DOI, &res.PageNumStart,
			&res.PageNumEnd, &res.Year)
		if err != nil {
			slog.Error("Cannot read page row", "page_id", pageID, "err", err)
			return nil, err
		}
		parts = append(parts, res)
	}
	if len(parts) == 0 {
		return nil, nil
	}
	return &parts[len(parts)-1], nil
}

func (l reffndio) nameOnlyOccurrences(nameRefs *bhl.RefsByName) ([]*refRec, error) {
	return l.occurrences(nameRefs.Canonical, "matched_canonical")
}

func (l reffndio) taxonOccurrences(nameRefs *bhl.RefsByName) ([]*refRec, error) {
	return l.occurrences(nameRefs.CurrentCanonical, "current_canonical")
}

func (l reffndio) occurrences(name string, field string) ([]*refRec, error) {
	switch field {
	case "matched_canonical", "current_canonical":
	default:
		slog.Warn("Unregistered field", "field", field)
		return nil, nil
	}
	var res []*refRec
	var itemID, titleID, pageID int
	var kingdomPercent, pathsTotal, editDistance sql.NullInt16
	var yearStart, yearEnd, titleYearStart, titleYearEnd sql.NullInt32
	var pageNum sql.NullInt64
	var nameID string
	var titleName, contextWrds, majorKingdom, nameString, matchedCanonical,
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

	ctx := context.Background()
	rows, err := l.db.Query(ctx, q, name)
	if err != nil {
		slog.Error("Cannot run occurences query", "error", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&itemID, &titleID, &pageID, &pageNum, &annot,
			&titleYearStart, &titleYearEnd, &yearStart, &yearEnd, &titleName, &vol,
			&titleDOI, &contextWrds, &majorKingdom, &kingdomPercent, &pathsTotal,
			&nameID, &nameString, &matchedCanonical, &matchType, &editDistance)
		if err != nil {
			err = fmt.Errorf("reffinderio.occurrences: %w", err)
			slog.Error("Cannot scan row", "error", err)
			return nil, err
		}
		rec := &refRec{
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
			mainTaxon:          contextWrds.String,
			mainKingdom:        majorKingdom.String,
			mainKingdomPercent: int(kingdomPercent.Int16),
			namesTotal:         int(pathsTotal.Int16),
			nameID:             nameID,
			name:               nameString.String,
			annotation:         annot.String,
			matchedCanonical:   matchedCanonical.String,
			matchType:          matchType.String,
			editDistance:       int(editDistance.Int16),
		}

		res = append(res, rec)
	}
	return res, nil
}

func (rf reffndio) currentCanonical(canonical string) (string, error) {
	var currentCan sql.NullString
	q := `SELECT current_canonical
          FROM name_strings
        WHERE matched_canonical = $1
        LIMIT 1`
	ctx := context.Background()
	row := rf.db.QueryRow(ctx, q, canonical)
	err := row.Scan(&currentCan)
	if err == pgx.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return currentCan.String, nil
}

func (rf *reffndio) colNomen(inp input.Input) (*bhl.RefsByName, error) {
	q := `
SELECT cr.result 
	FROM col_names cn
		JOIN col_bhl_results cr
			ON cn.id = cr.col_name_id
	WHERE cn.canonical_simple = $1
`

	ctx := context.Background()
	enc := gnfmt.GNgob{}
	var res []*bhl.RefsByName

	rows, err := rf.db.Query(ctx, q, inp.Name.CanonicalSimple)
	if err != nil {
		slog.Error("Cannot run CoL nomen query", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bs []byte
		var nref bhl.RefsByName
		err := rows.Scan(&bs)
		if err != nil {
			slog.Error("Cannot scan results of CoL nomen query", "error", err)
			return nil, err
		}

		err = enc.Decode(bs, &nref)
		if err != nil {
			slog.Error("Cannot decode name-reference data from CoL", "error", err)
			return nil, err
		}

		if len(nref.References) > 0 {
			res = append(res, &nref)
		}
	}
	var out *bhl.RefsByName
	switch len(res) {
	case 0:
		return nil, nil
	case 1:
		out = res[0]
	default:
		slices.SortFunc(res, func(a, b *bhl.RefsByName) int {
			return cmp.Compare(b.References[0].Odds, a.References[0].Odds)
		})
		out = res[0]
	}

	prepareOutput(inp, out)
	return out, nil
}

func prepareOutput(inp input.Input, nr *bhl.RefsByName) {
	nr.Input = inp
}
