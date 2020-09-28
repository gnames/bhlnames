package bhl

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cloudfoundry/bytefmt"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/db"
	"github.com/gnames/gnames/lib/sys"
	"github.com/gosuri/uiprogress"
)

type MetaData struct {
	config.Config
	DownloadFile string
	DownloadDir  string
	KeyValDir    string
	PartDir      string
	DB           *sql.DB
}

func NewMetaData(cfg config.Config) MetaData {
	md := MetaData{Config: cfg}
	md.DownloadFile = filepath.Join(cfg.InputDir, "data.zip")
	md.DownloadDir = filepath.Join(cfg.InputDir, "Data")
	md.KeyValDir = filepath.Join(cfg.InputDir, "keyval")
	md.PartDir = filepath.Join(cfg.InputDir, "part")
	return md
}

// Download will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory. We pass an
// io.TeeReader into Copy() to report progress on the download.
func (md MetaData) Download() error {
	path := md.DownloadFile
	if !md.Rebuild && sys.FileExists(path) {
		log.Printf("File %s already exists, skipping download.\n", path)
		return nil
	}
	out, err := os.Create(path + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(md.DumpURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	total := 0
	if size, ok := resp.Header["Content-Length"]; ok && len(size) > 0 {
		total, err = strconv.Atoi(size[0])
		if err != nil {
			return err
		}
	} else {
		return errors.New("cannot receive remote header of BHL data URL")
	}
	log.Printf(`Downloading %s of BHL data dump.`,
		bytefmt.ByteSize(uint64(total)))

	uiprogress.Start()
	counter := NewWriteCounter(total)
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}
	uiprogress.Stop()

	err = os.Rename(path+".tmp", path)
	if err != nil {
		return err
	}
	log.Println("Download finished")
	return nil
}

func (md MetaData) Extract() error {
	err := md.unzip()
	if err != nil {
		return err
	}
	return nil
}

func (md MetaData) Prepare() error {
	md.DB = db.NewDb(md.Config.DB)

	titleDOImap, partDOImap, err := md.prepareDOI()
	if err != nil {
		return err
	}
	titlesMap, err := md.prepareTitle(titleDOImap)
	if err != nil {
		return err
	}
	err = md.uploadItem(titlesMap)
	if err != nil {
		return err
	}
	err = md.uploadPart(partDOImap)
	if err != nil {
		return err
	}
	err = md.uploadPage()
	if err != nil {
		return err
	}

	return md.DB.Close()
}
