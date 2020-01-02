package bhl

import (
	"archive/zip"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bytefmt"
	"github.com/gnames/bhlnames/sys"
)

var files = map[string]struct{}{
	"Data/doi.txt":   struct{}{},
	"Data/item.txt":  struct{}{},
	"Data/page.txt":  struct{}{},
	"Data/part.txt":  struct{}{},
	"Data/title.txt": struct{}{},
}

func (md MetaData) unzip() error {
	if !sys.FileExists(md.DownloadFile) {
		return errors.New("cannot find BHL data dump file")
	}
	err := sys.MakeDir(md.DownloadDir)
	if err != nil {
		return err
	}

	err = sys.MakeDir(md.KeyValDir)
	if err != nil {
		return err
	}

	r, err := zip.OpenReader(md.DownloadFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if _, ok := files[f.Name]; !ok {
			continue
		}
		fpath := filepath.Join(md.InputDir, f.Name)
		if !md.Rebuild && sys.FileExists(fpath) {
			log.Printf("File %s already exists, skipping unzip", fpath)
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
		size := f.FileHeader.UncompressedSize64
		log.Printf("Extracting %s (%s)\n", f.Name, bytefmt.ByteSize(size))
		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
