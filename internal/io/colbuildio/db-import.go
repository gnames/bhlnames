package colbuildio

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/gnames/gnparser"
	"github.com/lib/pq"
)

func (c colbuildio) checkData() (bool, error) {
	var err error
	var hasTables bool
	var str string
	q := `
SELECT EXISTS (
    SELECT FROM
        pg_tables
    WHERE
        schemaname = 'public' AND
        tablename  = 'col_nomen_refs'
    )
`
	err = c.db.QueryRow(q).Scan(&hasTables)
	if !hasTables || err != nil {
		return hasTables, err
	}

	err = c.db.QueryRow(`SELECT record_id FROM col_nomen_refs limit 1`).Scan(&str)
	switch err {
	case sql.ErrNoRows:
		return false, nil
	case nil:
		return true, nil
	default:
		return false, err
	}
}

func (c colbuildio) saveNomenRefs(chIn <-chan []db.ColNomenRef) error {
	total := 0
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
	gnp := gnparser.New(gnparser.NewConfig())

	for refs := range chIn {
		total += len(refs)
		transaction, err := c.db.Begin()
		if err != nil {
			return err
		}

		// Find canonical forms
		names := make([]string, len(refs))
		for i := range refs {
			names[i] = refs[i].Name
		}
		ps := gnp.ParseNames(names)

		// Get classifications

		cls, err := c.classifications(refs)
		if err != nil {
			return err
		}

		stmt, err := transaction.Prepare(pq.CopyIn("col_nomen_refs", columns...))
		if err != nil {
			return err
		}

		for i, v := range refs {
			var can, canStem string
			if ps[i].Parsed {
				can = ps[i].Canonical.Simple
				canStem = ps[i].Canonical.Stemmed
			}
			cl := cls[v.RecordID]
			if cl["kingdom"] == "" {
				continue
			}

			_, err = stmt.Exec(
				v.RecordID, v.Name, v.Ref, cl["kingdom"], cl["phylum"], cl["class"],
				cl["order"], cl["family"], cl["genus"], can, canStem,
			)
			if err != nil {
				return err
			}
		}
		err = stmt.Close()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 47))
		fmt.Fprintf(os.Stderr, "\rProcessed %s CoL nomen refs", humanize.Comma(int64(total)))
		err = transaction.Commit()
		if err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr)

	return nil
}

func (c colbuildio) classifications(
	refs []db.ColNomenRef,
) (map[string]map[string]string, error) {
	var err error
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
	rows, err := c.db.Query(q, pq.Array(ids))
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
