package colbuildio

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gnames/bhlnames/io/db"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	refsBatchSize = 50_000
)

func (c colbuildio) importCol() error {
	var err error

	log.Info().Msg("Importing nomenclatural references from CoL.")
	log.Info().Msg("Truncating data from col_nomen_refs table.")
	c.resetColDB()

	g := errgroup.Group{}

	chRefs := make(chan []db.ColNomenRef)

	g.Go(func() error {
		err = c.saveNomenRefs(chRefs)
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

func (c colbuildio) loadNomenRefs(chRefs chan<- []db.ColNomenRef) error {
	path := c.pathExtract
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

	chunk := make([]db.ColNomenRef, 0, refsBatchSize)
	var count int

	for scan.Scan() {
		row := scan.Text()
		fields := strings.Split(row, "\t")
		if count == refsBatchSize {
			chRefs <- chunk
			chunk = make([]db.ColNomenRef, 0, refsBatchSize)
			count = 0
		}
		nRef := db.ColNomenRef{
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
