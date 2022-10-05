package colbuildio

import (
	"database/sql"

	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/io/db"
	"github.com/lib/pq"
)

const batchCOL = 10_000

func (c colbuildio) loadColData() ([]db.ColNomenRef, error) {
	res := make([]db.ColNomenRef, 0, batchCOL)
	q := `
SELECT record_id, name, ref FROM col_nomen_refs
  WHERE quality is NULL
  LIMIT $1
`
	rows, err := c.db.Query(q, batchCOL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cnr db.ColNomenRef
		err = rows.Scan(&cnr.RecordID, &cnr.Name, &cnr.Ref)
		if err != nil {
			return nil, err
		}
		res = append(res, cnr)
	}
	return res, nil
}

func (c colbuildio) updateColNomenRef(nrs *namerefs.NameRefs) error {
	var err error

	q := `
	UPDATE col_nomen_refs
	  SET page_id = $1, part_id = $2, item_id = $3,
	      refs_num = $4, odds = $5, quality = $6
	  WHERE record_id = $7
	`
	var pageID, partID, itemID, refsNum int
	var odds float64
	if len(nrs.References) > 0 {
		ref := nrs.References[0]
		pageID = ref.PageID
		partID = ref.PartID
		itemID = ref.ItemID
		refsNum = len(nrs.References)
		odds = ref.Score.Odds
	}
	quality := calcQuality(odds)
	_, err = c.db.Exec(q,
		pageID, partID, itemID,
		refsNum, odds, quality,
		nrs.Input.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c colbuildio) saveColNomenRef(nrs *namerefs.NameRefs) error {
	var err error
	var stmt *sql.Stmt
	var transaction *sql.Tx
	columns := []string{"record_id", "item_id", "part_id", "page_id", "odds", "quality"}

	for _, ref := range nrs.References {
		transaction, err = c.db.Begin()
		if err != nil {
			return err
		}
		stmt, err = transaction.Prepare(pq.CopyIn("col_bhl_refs", columns...))
		if err != nil {
			return err
		}

		_, err = stmt.Exec(nrs.Input.ID, ref.ItemID, ref.PartID, ref.PageID,
			ref.Score.Odds, calcQuality(ref.Score.Odds))
		if err != nil {
			return err
		}
		err = stmt.Close()
		if err != nil {
			return err
		}
		err = transaction.Commit()
		if err != nil {
			return err
		}
	}
	return nil
}

func calcQuality(odds float64) int {
	res := 0
	switch {
	case odds > 10:
		res = 4
	case odds > 1:
		res = 3
	case odds > 0.1:
		res = 2
	case odds > 0.01:
		res = 1
	}
	return res
}
