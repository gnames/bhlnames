package builderio

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/ent/txstats"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

func (b builderio) CalculateTxStats() error {
	var err error
	var itemsTotal int
	var itx []txstats.ItemTaxa
	log.Info().Msg("Calclulating taxonomic statistics for items.")
	itemsTotal, err = b.itemsNum()
	if err != nil {
		return fmt.Errorf("calculateStats: %w", err)
	}
	chIn := make(chan []txstats.ItemTaxa)
	eg := errgroup.Group{}

	eg.Go(func() error {
		return b.addStatsToItems(chIn)
	})

	itemID := 1
	limit := 1000
	var count int
	for itemID <= itemsTotal {
		itx, err = b.getItemsTaxa(itemID, limit)
		if err != nil {
			err = fmt.Errorf("calculateStats: %w", err)
			return err
		}
		chIn <- itx
		count += limit
		itemID += limit
		if count%25_000 == 0 {
			fmt.Fprint(os.Stderr, "\r")
			log.Info().Msgf("Calculated stats for %s items", humanize.Comma(int64(count)))
		} else {
			fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 35))
			fmt.Fprintf(os.Stderr, "\rCalculated taxonomic stats for %s items",
				humanize.Comma(int64(count)))
		}

	}
	close(chIn)

	if err = eg.Wait(); err != nil {
		err = fmt.Errorf("calculateStats: %w", err)
		return err
	}
	fmt.Println()
	log.Info().Msg("Finished calculation of taxonomic stats for items.")
	return nil
}
