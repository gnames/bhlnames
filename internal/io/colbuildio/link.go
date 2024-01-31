package colbuildio

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/gnparser"
)

func (c colbuildio) inputFromCol(chIn chan<- input.Input) error {
	slog.Info("Finding nomenclatural events for names from the Catalogue of Life.")

	gnp := gnparser.New(gnparser.NewConfig())
	cursor := c.lastProcRec
	for {
		cnr, err := c.loadColData(cursor)
		if err != nil {
			return err
		}
		cursor += batchCOL
		if len(cnr) == 0 {
			break
		}

		for i := range cnr {
			opts := []input.Option{
				input.OptID(cnr[i].RecordID),
				input.OptNameString(cnr[i].Name),
				input.OptRefString(cnr[i].Ref),
			}
			chIn <- input.New(gnp, opts...)
		}
	}
	return nil
}

func (c colbuildio) saveColBhlNomen(nrs *namerefs.NameRefs) error {
	var err error
	var transaction *sql.Tx

	transaction, err = c.db.Begin()
	if err != nil {
		return err
	}

	err = c.updateColNomenRef(nrs)
	if err != nil {
		err = fmt.Errorf("updateColNomenRef: %w", err)
		return err
	}

	err = c.saveColNomenRef(nrs, transaction)
	if err != nil {
		err = fmt.Errorf("updateColNomenRef: %w", err)
		return err
	}

	err = transaction.Commit()
	if err != nil {
		return err
	}

	return nil
}
