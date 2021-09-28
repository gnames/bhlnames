package builderio

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"code.cloudfoundry.org/bytefmt"
	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/gnsys"
	"github.com/gosuri/uiprogress"
)

// downloadDumpBHL will download a dump of BHL metadata using provided URL to a
// local file. It's efficient because it will write as it downloads and not
// load the whole file into memory. We pass an io.TeeReader into Copy() to
// report progress on the download.
func (b builderio) downloadDumpBHL() error {
	path := b.DownloadFile
	exists, _ := gnsys.FileExists(path)
	if !b.WithRebuild && exists {
		log.Printf("File %s already exists, skipping download.", path)
		return nil
	}
	out, err := os.Create(path + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(b.BHLDumpURL)
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

func (b builderio) uploadDataBHL() error {
	var err error
	var titlesMap map[int]*title
	var partDOImap map[int]string

	db.TruncateBHL(b.DB)
	titleDOImap, partDOImap, err := b.prepareDOI()

	if err == nil {
		titlesMap, err = b.prepareTitle(titleDOImap)
	}

	if err == nil {
		err = b.uploadItem(titlesMap)
	}

	if err == nil {
		err = b.uploadPart(partDOImap)
	}

	if err == nil {
		err = b.uploadPage()
	}

	if err == nil {
		tf := NewTitleFinder(b.Config, titlesMap)
		return tf.Setup()
	}

	return err
}
