package colbuildio

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/colbuild"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/io/bhlsys"
	"github.com/gnames/bhlnames/io/db"
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type colbuildio struct {
	dlURL        string
	dlDir        string
	pathDownload string
	pathExtract  string
	db           *sql.DB
	gormDB       *gorm.DB
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
	return err
}

func (c colbuildio) LinkColToBhl(nomenRef func(<-chan input.Input, chan<- *namerefs.NameRefs)) error {
	log.Info().Msg("Linking CoL references to BHL pages.")
	log.Warn().Msg("This part might take a few days.")

	var err error
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
	g2.Go(func() error {
		var count int
		for nrs := range chOut {
			count++
			err = c.saveColBhlNomen(nrs)
			if err != nil {
				err = fmt.Errorf("SaveColBhlNomen: %w", err)
				return err
			}
			fmt.Fprintf(os.Stderr, "\r%s", strings.Repeat(" ", 35))
			fmt.Fprintf(os.Stderr, "\rLinked %s CoL refs to BHL.", humanize.Comma(int64(count)))
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
	close(chOut)

	log.Info().Msg("Finished linking CoL nomenclatural references to BHL.")
	return g2.Wait()
}
