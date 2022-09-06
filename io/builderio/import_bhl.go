package builderio

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/gnsys"
	"github.com/rs/zerolog/log"
	progressbar "github.com/schollz/progressbar/v3"
)

// download will download a dump of BHL metadata using provided URL to a
// local file. It's efficient because it will write as it downloads and not
// load the whole file into memory. We pass an io.TeeReader into Copy() to
// report progress on the download.
func (b builderio) download(path, url string) error {
	exists, _ := gnsys.FileExists(path)
	if !b.WithRebuild && exists {
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

func (b builderio) importDataBHL() error {
	var err error
	var titlesMap map[int]*title
	var partDOImap map[int]string
	var itemMap map[uint]string
	var titleDOImap map[int]string

	err = db.Truncate(b.DB, []string{"items", "pages", "parts"})

	if err == nil {
		titleDOImap, partDOImap, err = b.prepareDOI()
	}

	if err == nil {
		titlesMap, err = b.prepareTitle(titleDOImap)
	}

	if err == nil {
		itemMap, err = b.importItem(titlesMap)
	}

	if err == nil {
		err = b.importPart(partDOImap)
	}

	if err == nil {
		err = b.importPage(itemMap)
	}

	if err == nil {
		ts := newTitleStore(b.Config, titlesMap)
		return ts.setup()
	}

	if err != nil {
		return fmt.Errorf("import BHL data: %w", err)
	}

	return nil
}
