package namesio

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	humanize "github.com/dustin/go-humanize"
	"github.com/gnames/bhlindex/protob"
	"github.com/gnames/bhlnames/io/db"
	"github.com/gnames/bhlnames/io/rpc"
	"github.com/gnames/gnparser"
	"github.com/gnames/uuid5"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type Names struct {
	InputDir  string
	HostRPC   string
	DB        *sql.DB
	GormDB    *gorm.DB
	BatchSize int
}

func New(host string, inputDir string) Names {

	n := Names{
		InputDir:  inputDir,
		HostRPC:   host,
		BatchSize: 100_000,
	}
	return n
}

func (n Names) ImportNames() error {
	log.Println("Truncating data from names_strings table.")
	err := db.TruncateNames(n.DB)
	if err != nil {
		return err
	}
	log.Println("Uploading name strings data via gRPC.")
	ch := make(chan []*protob.NameString)
	var wg sync.WaitGroup
	wg.Add(1)
	names := make([]*protob.NameString, 0, n.BatchSize)
	r := rpc.ClientRPC{Host: n.HostRPC}
	err = r.Connect()
	if err != nil {
		return err
	}

	stream, err := r.Client.Names(context.Background(), &protob.NamesOpt{})
	if err != nil {
		return err
	}

	go n.uploadNames(ch, &wg)
	namesNum := 0
	for {
		name, err := stream.Recv()
		namesNum++
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		names = append(names, name)
		if namesNum >= n.BatchSize {
			ch <- names
			names = make([]*protob.NameString, 0, n.BatchSize)
			namesNum = 0
		}
	}
	ch <- names
	close(ch)
	wg.Wait()
	return r.Conn.Close()
}

func (n Names) uploadNames(ch <-chan []*protob.NameString, wg *sync.WaitGroup) {
	defer wg.Done()
	cfg := gnparser.NewConfig()
	gnp := gnparser.New(cfg)
	total := 0
	namesUUID := make(map[string]struct{})

	for names := range ch {
		total += len(names)
		columns := []string{"id", "name", "taxon_id", "match_type",
			"edit_distance", "stem_edit_distance", "matched_name", "matched_canonical",
			"current_name", "current_canonical", "classification", "data_source_id",
			"data_source_title", "data_sources_number", "curation", "occurences",
			"odds", "error"}
		transaction, err := n.DB.Begin()
		if err != nil {
			log.Fatal(err)
		}
		stmt, err := transaction.Prepare(pq.CopyIn("name_strings", columns...))
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range names {
			id := uuid5.UUID5(v.Value).String()
			if _, ok := namesUUID[id]; ok {
				fmt.Printf("%s\n", v.Value)
				continue
			}
			namesUUID[id] = struct{}{}
			currentCanonical := ""
			if v.Current != "" {
				if v.Matched != v.Current {
					parsed := gnp.ParseName(v.Current)
					if parsed.Parsed {
						currentCanonical = parsed.Canonical.Full
					}
				} else {
					currentCanonical = v.MatchedCanonical
				}
			} else {
				v.Current = v.Matched
				currentCanonical = v.MatchedCanonical
			}

			dataSourceID := sql.NullInt64{}
			if v.DataSourceId > 0 {
				dataSourceID.Int64 = int64(v.DataSourceId)
				dataSourceID.Valid = true
			}

			_, err = stmt.Exec(id, v.Value, v.TaxonId, v.Match.String(),
				v.EditDistance, v.EditDistanceStem, v.Matched,
				v.MatchedCanonical, v.Current, currentCanonical, v.Classification,
				dataSourceID, v.DataSourceTitle, v.DataSourcesNum, v.Curated,
				v.Occurences, v.Odds, v.VerifError)
			if err != nil {
				log.Fatal(err)
			}
		}
		err = stmt.Close()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\r%s", strings.Repeat(" ", 35))
		fmt.Printf("\rUploaded %s names to db", humanize.Comma(int64(total)))
		err = transaction.Commit()
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println()
}
