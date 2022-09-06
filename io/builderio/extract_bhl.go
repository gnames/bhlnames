package builderio

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/bytefmt"
	"github.com/gnames/gnsys"
	"github.com/rs/zerolog/log"
)

var files = map[string]struct{}{
	"Data/doi.txt":    {},
	"Data/item.txt":   {},
	"Data/page.txt":   {},
	"Data/part.txt":   {},
	"Data/title.txt":  {},
	"occurrences.csv": {},
	"names.csv":       {},
	"pages.csv":       {},
}

func (b builderio) extract(path string) error {
	exists, _ := gnsys.FileExists(path)
	if !exists {
		return errors.New("cannot find BHL data dump file")
	}
	dataPath := b.DownloadDir
	exists, _, _ = gnsys.DirExists(dataPath)
	if !exists {
		err := gnsys.MakeDir(dataPath)
		if err != nil {
			return err
		}
	}
	return b.unzip(path)
}

func (b *builderio) unzip(path string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}

	defer r.Close()

	for _, f := range r.File {
		if _, ok := files[f.Name]; !ok {
			continue
		}
		fpath := filepath.Join(b.DownloadDir, filepath.Base(f.Name))
		exists, _ := gnsys.FileExists(fpath)
		if !b.WithRebuild && exists {
			log.Info().Msgf("File %s already exists, skipping unzip.", fpath)
			continue
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		size := f.UncompressedSize64
		log.Info().Msgf("Extracting %s (%s).", f.Name, bytefmt.ByteSize(size))
		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
