package builderio

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/internal/io/dbio"
	"golang.org/x/sync/errgroup"
)

type pagePart struct {
	pageID, partID int
}

func (b builderio) assignPartsToPages() error {
	slog.Info("Assigning parts IDs to pages IDs")
	ch := make(chan []pagePart)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := b.insertPagePart(ctx, ch)
		if err != nil {
			close(ch)
			slog.Error("Cannot insert page/part", "error", err)
			return err
		}
		return nil
	})

	err := b.loadPagePart(ctx, ch)
	if err != nil {
		close(ch)
		slog.Error("Cannot load page/part", "error", err)
		return err
	}
	close(ch)

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func (b builderio) loadPagePart(ctx context.Context, ch chan []pagePart) error {
	maxID, err := b.maxItemID()
	if err != nil {
		return err
	}
	curItemID := 1
	limit := 5000
	q := `
WITH subq AS (
  SELECT pt.id as part_id, pt.page_id, pt.length, pg.item_id, pg.sequence_order
    FROM parts pt
      JOIN pages pg ON pt.page_id = pg.id
    )
SELECT pg2.id as page_id, subq.part_id
  FROM pages pg2
    JOIN subq ON pg2.item_id = subq.item_id
  WHERE pg2.sequence_order >= subq.sequence_order
    AND pg2.sequence_order < subq.sequence_order + subq.length + 1
	  AND pg2.item_id >= $1
	  AND pg2.item_id < $2
  ORDER BY subq.part_id, pg2.sequence_order;
`
	for curItemID <= maxID {
		nextItemID := curItemID + limit
		rows, err := b.db.Query(
			context.Background(),
			q, curItemID, nextItemID,
		)
		if err != nil {
			return err
		}
		curItemID = nextItemID
		defer rows.Close()

		var pages []pagePart
		for rows.Next() {
			var p pagePart
			err = rows.Scan(&p.pageID, &p.partID)
			if err != nil {
				slog.Error("Cannot scan page/part", "error", err)
				return err
			}
			pages = append(pages, p)
		}

		select {

		case <-ctx.Done():
			return ctx.Err()
		default:
			ch <- pages
		}
	}
	return nil
}

func (b builderio) insertPagePart(ctx context.Context, ch chan []pagePart) error {
	columns := []string{"page_id", "part_id"}
	var count int
	for v := range ch {
		count += len(v)
		rows := make([][]any, len(v))
		for i := range v {
			row := []any{v[i].pageID, v[i].partID}
			rows[i] = row
		}

		_, err := dbio.InsertRows(b.db, "page_parts", columns, rows)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 35))
			fmt.Fprintf(os.Stderr, "\rProcessed %s pages/parts", humanize.Comma(int64(count)))
		}
	}

	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 35))
	slog.Info("Imported page/part data to db.", "records-num", humanize.Comma(int64(count)))
	return nil
}
