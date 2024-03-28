package builderio

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/internal/ent/txstats"
	"github.com/gnames/bhlnames/internal/io/dbio"
	"golang.org/x/sync/errgroup"
)

func (b builderio) CalculateTxStats() error {
	var err error
	var maxID int
	var itx []txstats.ItemTaxa
	slog.Info("Truncating item_stats table.")
	dbio.Truncate(b.db, []string{"item_stats"})

	maxID, err = b.maxItemID()
	if err != nil {
		return err
	}

	slog.Info("Calclulating taxonomic statistics for items.")
	chIn := make(chan []txstats.ItemTaxa)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return b.addStatsToItems(ctx, chIn)
	})

	itemID := 1
	limit := 5000
	var count int
	for itemID <= maxID {
		itx, err = b.getItemsTaxa(itemID, limit)
		if err != nil {
			for range chIn {
			}
			slog.Error("Failed to get items taxa.", "error", err)
			return err
		}

		chIn <- itx
		count += len(itx)
		itemID += limit

		fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 35))
		fmt.Fprintf(
			os.Stderr,
			"\rCalculated taxonomic stats for %s items.",
			humanize.Comma(int64(count)),
		)
	}
	close(chIn)

	if err = g.Wait(); err != nil {
		err = fmt.Errorf("calculateStats: %w", err)
		return err
	}

	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 35))
	slog.Info(
		"Calculated taxonomic stats for items.",
		"records-num", humanize.Comma(int64(count)),
	)
	return nil
}
