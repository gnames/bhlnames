package colbuildio

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/internal/ent/colbuild"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/bhlnames/internal/io/bhlsys"
	"github.com/gnames/bhlnames/internal/io/db"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnsys"
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type colbuildio struct {
	dlURL        string
	dlDir        string
	pathDownload string
	pathExtract  string

	recordsNum  int
	lastProcRec int

	db     *sql.DB
	gormDB *gorm.DB
}

func New(cfg config.Config) colbuild.ColBuild {
	res := colbuildio{
		dlURL:        cfg.CoLDataURL,
		dlDir:        cfg.DownloadDir,
		pathDownload: cfg.DownloadCoLFile,
		pathExtract:  filepath.Join(cfg.DownloadDir, "Taxon.tsv"),
		db:           db.NewDB(cfg),
		gormDB:       db.NewDbGorm(cfg),
	}
	return &res
}

func (c colbuildio) DataStatus() (bool, bool, error) {
	var err error
	var exists, hasFiles, hasData bool

	exists, _ = gnsys.FileExists(c.pathDownload)
	if exists {
		exists, _ = gnsys.FileExists(c.pathExtract)
	}
	if exists {
		hasFiles = true
	} else {
		return hasFiles, hasData, err
	}

	hasData, err = c.checkData()
	return hasFiles, hasData, err
}

func (c colbuildio) ResetColData() {
	log.Info().Msg("Reseting CoL files")
	c.deleteFiles()
	c.resetColDB()
}

func (c colbuildio) ImportColData() error {
	var err error
	log.Info().Msg("Downloading CoL DwCA data.")
	err = bhlsys.Download(c.pathDownload, c.dlURL, false)
	if err != nil {
		err = fmt.Errorf("download: %w", err)
		return err
	}

	err = bhlsys.Extract(c.pathDownload, c.dlDir, false)
	if err != nil {
		err = fmt.Errorf("extract: %w", err)
		return err
	}

	err = c.importCol()
	if err != nil {
		err = fmt.Errorf("importCol: %w", err)
		return err
	}
	return nil
}

func (c colbuildio) LinkColToBhl(nomenRef func(<-chan input.Input, chan<- *namerefs.NameRefs)) error {
	var err error
	log.Info().Msg("Linking CoL references to BHL pages.")
	log.Warn().Msg("This part might take a few days.")

	c.recordsNum, c.lastProcRec, err = c.stats()
	if err != nil {
		return err
	}

	log.Info().Msgf("Processing %d CoL records.", c.recordsNum)

	if c.lastProcRec > 0 {
		log.Info().Msgf("Already processed %d records out of %d (%0.2f%%).",
			c.lastProcRec, c.recordsNum, c.recNumToPcent(c.lastProcRec))
	}

	start := time.Now()
	chIn := make(chan input.Input)
	chOut := make(chan *namerefs.NameRefs)

	g1 := errgroup.Group{}
	g2 := errgroup.Group{}

	// find references
	g1.Go(func() error {
		nomenRef(chIn, chOut)
		return nil
	})

	// save references
	var count int
	g2.Go(func() error {
		for nrs := range chOut {
			count++
			err = c.saveColBhlNomen(nrs)
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
		err = fmt.Errorf("InputFromCoL: %w", err)
		return err
	}
	close(chIn)

	if err = g1.Wait(); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr)
	log.Info().Msg("Finished linking CoL nomenclatural references to BHL.")
	err = g2.Wait()
	if err != nil {
		return err
	}
	dur := float64(time.Since(start)) / float64(time.Hour)
	log.Info().Msgf("Processed %d records in %0.2f hours.", count, dur)
	return nil
}

func (c colbuildio) recNumToPcent(recNum int) float64 {
	return float64(recNum) / float64(c.recordsNum) * 100
}

func (c colbuildio) progressOutput(start time.Time, recsNum int) {
	dur := float64(time.Since(start)) / float64(time.Second)
	rate := float64(recsNum-c.lastProcRec) / dur * 3600

	rateStr := humanize.Comma(int64(rate))
	eta := 3600 * float64(c.recordsNum-recsNum) / rate
	recs := humanize.Comma(int64(recsNum))

	str := fmt.Sprintf("Linked %s (%0.2f%%) CoL refs to BHL, %s rec/hr, ETA: %s",
		recs, c.recNumToPcent(recsNum), rateStr, gnfmt.TimeString(eta))
	fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 80))
	if recsNum%10_000 == 0 {
		fmt.Fprint(os.Stderr, "\r")
		log.Info().Msg(str)
	} else {
		fmt.Fprintf(os.Stderr, "\r%s", str)
	}
}
