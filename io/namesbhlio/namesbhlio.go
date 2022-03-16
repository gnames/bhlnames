package namesbhlio

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gnames/bhlindex/ent/name"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/namebhl"
	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/gnfmt"
	"github.com/jinzhu/gorm"
	"golang.org/x/sync/errgroup"
)

type namesbhlio struct {
	cfg    config.Config
	client *http.Client
	db     *sql.DB
	gormDB *gorm.DB
}

func New(cfg config.Config, db *sql.DB, gormdb *gorm.DB) namebhl.NameBHL {
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 10 * time.Second,
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	res := namesbhlio{cfg: cfg, client: client, db: db, gormDB: gormdb}
	return res
}

func (n namesbhlio) ImportOccurrences() error {
	log.Println("Ingesting names' occurrences.")
	log.Println("Truncating data from page_name_strings table.")
	err := db.Truncate(n.db, []string{"page_name_strings"})
	if err != nil {
		return fmt.Errorf("ImportOccurrences: %w", err)
	}

	kv := db.InitKeyVal(n.cfg.PageDir)
	defer kv.Close()
	g := errgroup.Group{}

	chOccur := make(chan []name.DetectedName)

	g.Go(func() error {
		return n.saveOcurrences(kv, chOccur)
	})

	offsetID := 1
	limit := 20_000
	var occurrs []name.DetectedName
	for {
		occurrs, err = n.getOccurrences(offsetID, limit)
		if err != nil {
			return fmt.Errorf("ImportOccurrences: %w", err)
		}
		if len(occurrs) == 0 {
			break
		}
		chOccur <- occurrs
		offsetID += limit
	}
	close(chOccur)

	err = g.Wait()
	if err != nil {
		return fmt.Errorf("ImportOccurrences: %w", err)
	}
	return nil
}

func (n namesbhlio) getREST(url string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Cannot create request: %v", err)
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(request)
	if err != nil {
		log.Print("Cannot get occurrences from BHLindex")
		return nil, err
	}
	defer resp.Body.Close()

	var respBytes []byte
	respBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBytes, err
}

func (n namesbhlio) getOccurrences(offsetID, limit int) ([]name.DetectedName, error) {
	namesURL := fmt.Sprintf("%soccurrences?limit=%d&offset_id=%d",
		n.cfg.BHLIndexURL, limit, offsetID)
	var res []name.DetectedName

	respBytes, err := n.getREST(namesURL)
	if err != nil {
		return nil, fmt.Errorf("getOccurrences: %w", err)
	}

	err = gnfmt.GNjson{}.Decode(respBytes, &res)
	if err != nil {
		return nil, fmt.Errorf("getOccurrences: %w", err)
	}
	return res, nil
}

func (n namesbhlio) ImportNames() error {
	log.Println("Ingesting names resolved to Catalogue of Life.")
	log.Println("Truncating data from names_strings table.")
	err := db.Truncate(n.db, []string{"name_strings"})
	if err != nil {
		return fmt.Errorf("import names: %w", err)
	}

	lastID, err := n.namesLastID()
	if err != nil {
		return fmt.Errorf("import names: %w", err)
	}

	log.Printf("Downloading BHL's verified names. LastID: %s", humanize.Comma(int64(lastID)))
	chNames := make(chan []name.VerifiedName)
	g := errgroup.Group{}

	g.Go(func() error {
		return n.saveNames(chNames)
	})

	offsetID := 1
	limit := 20_000
	for offsetID < lastID {
		enc := gnfmt.GNjson{}
		namesURL := fmt.Sprintf("%snames?data_sources=1&limit=%d&offset_id=%d",
			n.cfg.BHLIndexURL, limit, offsetID)
		var res []name.VerifiedName
		request, err := http.NewRequest(http.MethodGet, namesURL, nil)
		if err != nil {
			return fmt.Errorf("import names: %w", err)
		}
		request.Header.Set("Content-Type", "application/json")

		resp, err := n.client.Do(request)
		if err != nil {
			log.Print("Cannot get verified names from BHLindex.")
			return err
		}
		defer resp.Body.Close()

		var respBytes []byte
		respBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Print("Body reading is failing for a search")
			return err
		}
		err = enc.Decode(respBytes, &res)
		if err != nil {
			log.Print("Cannot decode search result")
			return err
		}
		chNames <- res
		offsetID += limit
	}
	close(chNames)
	return g.Wait()
}

func (n namesbhlio) namesLastID() (int, error) {
	var err error
	var res int
	url := fmt.Sprintf("%snames/last_id", n.cfg.BHLIndexURL)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Cannot create request: %v", err)
		return res, err
	}
	request.Header.Set("Content-Type", "text/plain")

	resp, err := n.client.Do(request)
	if err != nil {
		log.Print("Cannot get names last ID")
		return res, err
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}
	return strconv.Atoi(string(bs))
}
