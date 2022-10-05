package colbuildio

import (
	"fmt"

	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/gnparser"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

func (c colbuildio) inputFromCol(chIn chan<- input.Input) error {
	log.Info().Msg("Finding nomenclatural events for names from the Catalogue of Life.")
	log.Info().Msg("Truncating data from col_bhl_refs table.")

	err := db.Truncate(c.db, []string{"col_bhl_refs"})
	if err != nil {
		return err
	}

	gnp := gnparser.New(gnparser.NewConfig())
	for {
		cnr, err := c.loadColData()
		if err != nil {
			return err
		}

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
	g := errgroup.Group{}

	g.Go(func() error {
		err = c.updateColNomenRef(nrs)
		if err != nil {
			err = fmt.Errorf("updateColNomenRef: %w", err)
		}
		return err
	})

	err = c.saveColNomenRef(nrs)
	if err != nil {
		err = fmt.Errorf("updateColNomenRef: %w", err)
		return err
	}

	return g.Wait()
}
