package colio

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnames/bhlnames/internal/ent/model"
	"golang.org/x/sync/errgroup"
)

const (
	refsBatchSize = 50_000
)

func (c colio) importCoL() error {
	var err error

	slog.Info("Importing nomenclatural references from CoL.")
	slog.Info("Truncating data from col_names table.")
	c.resetColDB()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	chRefs := make(chan []model.ColName)

	g.Go(func() error {
		err = c.processColNames(ctx, chRefs)
		if err != nil {
			err = fmt.Errorf("saveNomenRefs: %w", err)
		}
		return err
	})

	err = c.loadNomenRefs(chRefs)
	if err != nil {
		return fmt.Errorf("loadNomenRefs: %w", err)
	}
	close(chRefs)

	return g.Wait()
}

const (
	colTaxonIDF = 0
	colSciNameF = 8
	colRefF     = 17
)

func (c colio) loadNomenRefs(chRefs chan<- []model.ColName) error {
	path := filepath.Join(c.cfg.ExtractDir, "Taxon.tsv")
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	scan := bufio.NewScanner(f)

	// some lines are too long for default 64k buffer
	maxCapacity := 300_000
	buf := make([]byte, maxCapacity)
	scan.Buffer(buf, maxCapacity)

	// skip headers
	scan.Scan()

	chunk := make([]model.ColName, 0, refsBatchSize)
	var count int

	for scan.Scan() {
		row := scan.Text()
		fields := strings.Split(row, "\t")
		if count == refsBatchSize {
			chRefs <- chunk
			chunk = make([]model.ColName, 0, refsBatchSize)
			count = 0
		}
		nRef := model.ColName{
			RecordID: fields[colTaxonIDF],
			Name:     fields[colSciNameF],
			Ref:      fields[colRefF],
		}
		if nRef.Ref != "" && len(nRef.Ref) < 2000 {
			chunk = append(chunk, nRef)
			count++
		}
	}

	chRefs <- chunk
	return scan.Err()
}
