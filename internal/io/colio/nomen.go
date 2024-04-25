package colio

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/gnfmt"
	"golang.org/x/sync/errgroup"
)

const (
	batchCOL = 100
	logBatch = 1_000_000
)

func (c *colio) NomenEvents(
	nomenRef func(
		context.Context,
		<-chan input.Input,
		chan<- *bhl.RefsByName,
	) error,
) error {
	var err error
	slog.Info("Linking CoL references to BHL pages.")
	slog.Warn("This part might take several hours.")

	c.recordsNum, c.lastProcRec, err = c.stats()
	if err != nil {
		return err
	}

	slog.Info("Processing CoL records.", "records-num", c.recordsNum)

	if c.lastProcRec > 0 {
		slog.Info(
			"Processing records",
			"records-num", c.lastProcRec,
			"all-records-num", c.recordsNum,
			"percent", c.recNumToPcent(c.lastProcRec),
		)
	}

	start := time.Now()
	chIn := make(chan input.Input)
	chOut := make(chan *bhl.RefsByName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		nomenRef(ctx, chIn, chOut)
		return nil
	})

	// save references
	var count int
	g.Go(func() error {
		for nrs := range chOut {
			count++
			err = c.saveColBhlRefs(nrs)
			if err != nil {
				err = fmt.Errorf("SaveColBhlNomen: %w", err)
				return err
			}
			recsNum := count + c.lastProcRec
			if recsNum%100 == 0 {
				c.progressOutput(start, recsNum)
			}
		}
		return nil
	})

	// load input data
	err = c.inputFromCol(chIn)
	if err != nil {
		slog.Error("Cannot generate input from CoL data.", "error", err)
		return err
	}
	close(chIn)

	if err = g.Wait(); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr)
	slog.Info("Finished linking CoL nomenclatural references to BHL.")

	dur := float64(time.Since(start)) / float64(time.Hour)
	durStr := fmt.Sprintf("%0.2f", dur)
	slog.Info("Stats", "records-num", count, "hours", durStr)
	return nil
}

func (c colio) recNumToPcent(recNum int) float64 {
	return float64(recNum) / float64(c.recordsNum) * 100
}

func (c colio) progressOutput(start time.Time, recsNum int) {
	dur := float64(time.Since(start)) / float64(time.Second)
	rate := float64(recsNum-c.lastProcRec) / dur * 3600

	rateStr := humanize.Comma(int64(rate))
	eta := 3600 * float64(c.recordsNum-recsNum) / rate
	recs := humanize.Comma(int64(recsNum))

	str := fmt.Sprintf("Linked %s (%0.2f%%) CoL refs to BHL, %s rec/hr, ETA: %s",
		recs, c.recNumToPcent(recsNum), rateStr, gnfmt.TimeString(eta))
	fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 80))
	if recsNum%logBatch == 0 {
		fmt.Fprint(os.Stderr, "\r")
		slog.Info(str)
	} else {
		fmt.Fprintf(os.Stderr, "\r%s", str)
	}
}

func (c colio) inputFromCol(chIn chan<- input.Input) error {
	slog.Info("Finding nomenclatural events for names from the Catalogue of Life.")

	gnp := <-c.gnpPool
	defer func() {
		c.gnpPool <- gnp
	}()
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
			id := strconv.Itoa(int(cnr[i].ID))
			opts := []input.Option{
				input.OptID(id + "|" + cnr[i].RecordID),
				input.OptNameString(cnr[i].Name),
				input.OptRefString(cnr[i].Ref),
				input.OptWithNomenEvent(true),
			}
			chIn <- input.New(c.gnpPool, opts...)
		}
	}
	return nil
}
