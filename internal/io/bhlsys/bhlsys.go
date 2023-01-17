package bhlsys

import (
	"archive/zip"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/bytefmt"
	"github.com/gnames/gnsys"
	"github.com/rs/zerolog/log"
	progressbar "github.com/schollz/progressbar/v3"
)

var files = map[string]struct{}{
	// BHL data files
	"Data/doi.txt":   {},
	"Data/item.txt":  {},
	"Data/page.txt":  {},
	"Data/part.txt":  {},
	"Data/title.txt": {},

	// BHLIndex files
	"occurrences.csv": {},
	"names.csv":       {},
	"pages.csv":       {},

	// CoL file
	"Taxon.tsv": {},
}

func Extract(path, dlDir string, rebuild bool) error {
	exists, _ := gnsys.FileExists(path)
	if !exists {
		return errors.New("cannot find BHL data dump file")
	}
	exists, _, _ = gnsys.DirExists(dlDir)
	if !exists {
		err := gnsys.MakeDir(dlDir)
		if err != nil {
			return err
		}
	}
	return unzip(path, dlDir, rebuild)
}

func unzip(path, dlDir string, rebuild bool) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}

	defer r.Close()

	for _, f := range r.File {
		if _, ok := files[f.Name]; !ok {
			continue
		}
		fpath := filepath.Join(dlDir, filepath.Base(f.Name))
		exists, _ := gnsys.FileExists(fpath)
		if !rebuild && exists {
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

// Download will Download a dump of BHL metadata using provided URL to a
// local file. It's efficient because it will write as it downloads and not
// load the whole file into memory. We pass an io.TeeReader into Copy() to
// report progress on the Download.
func Download(path, url string, rebuild bool) error {
	exists, _ := gnsys.FileExists(path)
	if !rebuild && exists {
		log.Info().Msgf("File %s already exists, skipping download.", path)
		return nil
	}
	out, err := os.Create(path + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)
	io.Copy(io.MultiWriter(out, bar), resp.Body)

	err = os.Rename(path+".tmp", path)
	if err != nil {
		return err
	}
	log.Info().Msg("Download finished")
	return nil
}