package colio

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/gnames/bhlnames/internal/io/dbio"
	"github.com/gnames/gnparser"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
)

func (c *colio) checkData() (bool, error) {
	ctx := context.Background()
	var err error
	var hasTables bool
	var str string
	q := `
SELECT EXISTS (
    SELECT FROM
        pg_tables
    WHERE
        schemaname = 'public' AND
        tablename  = 'col_names'
    )
`
	err = c.db.QueryRow(ctx, q).Scan(&hasTables)
	if !hasTables || err != nil {
		return hasTables, err
	}

	err = c.db.QueryRow(
		ctx, `SELECT record_id FROM col_names limit 1`,
	).Scan(&str)
	switch err {
	case pgx.ErrNoRows:
		return false, nil
	case nil:
		return true, nil
	default:
		return false, err
	}
}

func (c colio) processColNames(
	ctx context.Context,
	chIn <-chan []model.ColName,
) error {
	total := 0
	gnp := <-c.gnpPool
	defer func() {
		c.gnpPool <- gnp
	}()

	for refs := range chIn {
		total += len(refs)
		err := c.saveColNames(refs, gnp)
		if err != nil {
			slog.Error("Error saving nomen refs.", "error", err)
			return err
		}

		select {
		case <-ctx.Done():
			slog.Info("Importing nomen refs canceled.")
			return ctx.Err()
		default:
		}
		fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 35))
		fmt.Fprintf(
			os.Stderr,
			"\rImported %s names from CoL to db", humanize.Comma(int64(total)),
		)
	}
	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 35))
	slog.Info(
		"Finished importing data from CoL to the database",
		"records", humanize.Comma(int64(total)),
	)
	return nil
}

func (c colio) saveColNames(
	refs []model.ColName,
	gnp gnparser.GNparser) error {
	var can, canStem string
	columns := []string{
		"record_id",
		"name",
		"ref",
		"kingdom",
		"phylum",
		"class",
		"ordr",
		"family",
		"genus",
		"canonical_simple",
		"canonical_stem",
	}

	rows := make([][]any, 0, len(refs))
	names := make([]string, len(refs))
	for i := range refs {
		names[i] = refs[i].Name
	}
	ps := gnp.ParseNames(names)

	// Get classifications
	cls, err := c.classifications(refs)
	if err != nil {
		slog.Error("Error getting classifications.", "error", err)
		return err
	}

	for i, v := range refs {
		if ps[i].Parsed {
			can = ps[i].Canonical.Simple
			canStem = ps[i].Canonical.Stemmed
		}
		cl := cls[v.RecordID]
		row := []any{v.RecordID, v.Name, v.Ref, cl["kingdom"], cl["phylum"], cl["class"],
			cl["order"], cl["family"], cl["genus"], can, canStem,
		}
		rows = append(rows, row)
	}
	_, err = dbio.InsertRows(c.db, "col_names", columns, rows)
	if err != nil {
		return err
	}
	return nil
}

func (c colio) classifications(
	refs []model.ColName,
) (map[string]map[string]string, error) {
	var err error
	ctx := context.Background()
	var recID, clNames, clRanks string
	res := make(map[string]map[string]string)
	ids := make([]string, len(refs))
	for i := range refs {
		ids[i] = refs[i].RecordID
	}
	q := `
  SELECT record_id, classification, classification_ranks
    FROM name_strings
    WHERE record_id = any($1::varchar[])
    GROUP BY record_id, classification, classification_ranks
`
	rows, err := c.db.Query(ctx, q, pq.Array(ids))
	if err != nil {
		return res, err
	}
	defer rows.Close()
	var count int
	for rows.Next() {
		count++
		err := rows.Scan(&recID, &clNames, &clRanks)
		if err != nil {
			return res, err
		}
		res[recID] = classifRecord(clNames, clRanks)

	}
	var count2 int
	for _, v := range res {
		if strings.TrimSpace(v["kingdom"]) != "" {
			count2++
		}
	}
	return res, nil
}

func classifRecord(clNames, clRanks string) map[string]string {
	res := map[string]string{
		"kingdom": "", "phylum": "", "class": "",
		"order": "", "family": "", "genus": "",
	}

	names := strings.Split(clNames, "|")
	ranks := strings.Split(clRanks, "|")
	if len(names) < 2 || len(names) != len(ranks) {
		return res
	}

	for i, v := range ranks {
		if _, ok := res[v]; ok {
			res[v] = names[i]
		}
	}
	return res
}

func (c colio) stats() (int, int, error) {
	var num, numDone int
	cxt := context.Background()
	err := c.db.QueryRow(cxt, "SELECT MAX(id) FROM col_names").Scan(&num)

	switch err {
	case pgx.ErrNoRows, nil:
	default:
		return num, numDone, err
	}

	err = c.db.QueryRow(
		cxt,
		`SELECT col_name_id FROM col_bhl_refs LIMIT 1`,
	).Scan(&numDone)
	if err == pgx.ErrNoRows {
		return num, 0, nil
	}

	err = c.db.QueryRow(
		cxt,
		`SELECT max(col_name_id) FROM col_bhl_refs`,
	).Scan(&numDone)

	switch err {
	case pgx.ErrNoRows:
	case nil:
		numDone--
	default:
		return num, numDone, err
	}

	return num, numDone, nil
}

func (c colio) loadColData(offset int) ([]model.ColName, error) {
	ctx := context.Background()
	res := make([]model.ColName, 0, batchCOL)
	q := `
SELECT id, record_id, name, ref FROM col_names
  ORDER BY id
  OFFSET $1 
  LIMIT $2
`
	rows, err := c.db.Query(ctx, q, offset, batchCOL)
	if err != nil {
		slog.Error("Cannod run CoL data query", "error", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cnr model.ColName
		err = rows.Scan(&cnr.ID, &cnr.RecordID, &cnr.Name, &cnr.Ref)
		if err != nil {
			slog.Error("Cannot scan CoL data", "error", err)
			return nil, err
		}
		res = append(res, cnr)
	}
	return res, nil
}

func (c colio) saveColBhlRefs(refs *bhl.RefsByName) error {
	var err error
	columns := []string{
		"col_name_id",
		"record_id",
		"matched_name",
		"edit_distance",
		"annot_nomen",
		"item_id",
		"part_id",
		"page_id",
		"ref_match_quality",
		"score_odds",
		"score_total",
		"score_annot",
		"score_year",
		"score_ref_title",
		"score_ref_volume",
		"score_ref_pages",
	}
	rows := make([][]any, 0, len(refs.References))
	es := strings.Split(refs.Input.ID, "|")
	id, _ := strconv.Atoi(es[0])
	recID := es[1]
	for _, ref := range refs.References {

		row := []any{id, recID, ref.MatchedName, ref.EditDistance, ref.AnnotNomen,
			ref.ItemID, ref.Part.ID, ref.PageID, ref.RefMatchQuality,
			ref.Score.Odds, ref.Score.Total, ref.Score.Annot, ref.Score.Year,
			ref.Score.RefTitle, ref.Score.RefVolume, ref.Score.RefPages}

		rows = append(rows, row)
	}
	_, err = dbio.InsertRows(c.db, "col_bhl_refs", columns, rows)
	if err != nil {
		slog.Error("Cannot insert CoL ref data", "error", err)
		return err
	}
	return nil
}
