package builder_pg

import (
	"archive/zip"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bytefmt"
	"github.com/gnames/gnlib/sys"
)

var files = map[string]struct{}{
	"Data/doi.txt":   {},
	"Data/item.txt":  {},
	"Data/page.txt":  {},
	"Data/part.txt":  {},
	"Data/title.txt": {},
}

func (b BuilderPG) extractFilesBHL() error {
	if !sys.FileExists(b.Config.DownloadFile) {
		return errors.New("cannot find BHL data dump file")
	}
	r, err := zip.OpenReader(b.Config.DownloadFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if _, ok := files[f.Name]; !ok {
			continue
		}
		fpath := filepath.Join(b.Config.InputDir, f.Name)
		if !b.Config.Rebuild && sys.FileExists(fpath) {
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
