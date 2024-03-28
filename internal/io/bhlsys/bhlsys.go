package bhlsys

import (
	"archive/zip"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/bytefmt"
	"github.com/gnames/gnsys"
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

// Extract will extract files from a BHL data dump file into a directory.
// If rebuild is true, it will overwrite existing files, if false it will
// skip existing files.
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
			slog.Info("Skipping unzip, file already exists.", "file", fpath)
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
		slog.Info(
			"Extracting file from BHL dump.",
			"file", f.Name, "bytes-size", bytefmt.ByteSize(size),
		)
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
		slog.Info("File already exists, skipping download.", "file", path)
		return nil
	}

	dir := filepath.Dir(path)

	localPath, err := gnsys.Download(url, dir, true)
	if err != nil {
		return err
	}
	os.Rename(localPath, path)

	slog.Info("Download finished.")
	return nil
}
