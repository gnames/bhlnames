package names

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlindex/protob"
	"github.com/gnames/bhlnames/db"
	"github.com/gnames/bhlnames/rpc"
	"github.com/gnames/uuid5"
	"github.com/lib/pq"
)

func (n Names) ImportNamesOccur(keyValDir string) error {
	err := db.TruncateOccur(n.DB)
	if err != nil {
		return err
	}

	kv := db.InitKeyVal(keyValDir)
	defer kv.Close()

	r := rpc.ClientRPC{Host: n.HostRPC}
	err = r.Connect()
	if err != nil {
		return err
	}
	stream, err := r.Client.Pages(context.Background(), &protob.PagesOpt{})
	if err != nil {
		return err
	}
	err = n.getItems(kv, stream)
	if err != nil {
		return err
	}
	return r.Conn.Close()
}

func (n Names) getItems(kv *badger.DB, stream protob.BHLIndex_PagesClient) error {
	missingItems := make(map[string]struct{})
	defer func() {
		err := n.saveMissingItems(missingItems)
		if err != nil {
			log.Fatal(err)
		}
	}()
	item := &db.Item{}
	// collects classifications
	pathMap := make(map[string]struct{})
	var count, itemCount int

	chNames := make(chan *db.PageNameString)
	var wg sync.WaitGroup
	wg.Add(1)

	go n.saveOccurences(chNames, &wg)

	for {
		pageData, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		itemBarCode := pageData.ItemId

		if itemBarCode != item.BarCode {
			count = 0
			if item.BarCode != "" {
				itemCount++
				if itemCount%10_000 == 0 {
					fmt.Println()
					log.Printf("%d items processed\n", itemCount)
				}
				err := n.processItem(item, pathMap)
				if err != nil {
					return err
				}
				pathMap = make(map[string]struct{})
			}
			item = &db.Item{}
			n.GormDB.Where("bar_code = ?", itemBarCode).First(item)
		}

		count++
		pageSeqNum := count
		if item.ID == 0 {
			missingItems[itemBarCode] = struct{}{}
		}
		key := fmt.Sprintf("%d|%d", pageSeqNum, item.ID)
		pageID := db.GetValue(kv, key)
		if pageID == 0 {
			continue
		}
		for _, path := range n.processPage(pageData, pageID, chNames) {
			if _, ok := pathMap[path]; !ok {
				pathMap[path] = struct{}{}
			}
		}
	}
	close(chNames)
	wg.Wait()
	log.Println("Finished getting occurence data.")
	return nil
}

func (n Names) saveOccurences(ch <-chan *db.PageNameString,
	wg *sync.WaitGroup) {
	defer wg.Done()
	batch := 10_000
	var total, count int
	occurs := make([]*db.PageNameString, 0, batch)
	for occur := range ch {
		count++
		occurs = append(occurs, occur)
		if count >= batch {
			total += count
			count = 0
			err := n.uploadOccurences(occurs)
			if err != nil {
				log.Fatal(err)
			}
			occurs = make([]*db.PageNameString, 0, batch)
			fmt.Printf("\r%s", strings.Repeat(" ", 35))
			fmt.Printf("\rUploaded %d occurences to db", total)
		}
	}
	total += count
	err := n.uploadOccurences(occurs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rUploaded %d occurences to db", total)
	fmt.Println()
}

func (n Names) uploadOccurences(occurs []*db.PageNameString) error {
	columns := []string{"page_id", "name_string_id", "offset_start",
		"offset_end", "odds", "annotation", "annotation_type"}
	transaction, err := n.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := transaction.Prepare(pq.CopyIn("page_name_strings", columns...))
	if err != nil {
		return err
	}
	for _, v := range occurs {
		_, err = stmt.Exec(v.PageID, v.NameStringID, v.OffsetStart, v.OffsetEnd,
			v.Odds, v.Annotation, v.AnnotationType)
		if err != nil {
			return err
		}
	}
	err = stmt.Close()
	if err != nil {
		return err
	}
	return transaction.Commit()
}

func (n Names) saveMissingItems(data map[string]struct{}) error {
	log.Println("Writing down items not found in BarCode field of item.txt file.")
	path := filepath.Join(n.InputDir, "missing_items.txt")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	defer func() {
		err := f.Sync()
		if err != nil {
			log.Fatal(err)
		}
	}()
	_, err = f.WriteString("List of item BarCodes not found in item.txt\n")
	if err != nil {
		return err
	}
	for k := range data {
		_, err = f.WriteString(k + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

// process Page saves occurences that were verified by Catalogue of Life
func (n Names) processPage(pageData *protob.Page, pageID int,
	ch chan<- *db.PageNameString) []string {
	// classifications collections
	paths := make([]string, 0, len(pageData.Names))
	for _, n := range pageData.Names {
		if n.DataSourceId == 1 {
			paths = append(paths, n.Classification)
			pn := &db.PageNameString{
				PageID:         uint(pageID),
				NameStringID:   uuid5.UUID5(n.Value).String(),
				OffsetStart:    uint(n.OffsetStart),
				OffsetEnd:      uint(n.OffsetEnd),
				Odds:           float64(n.Odds),
				Annotation:     n.Annotation,
				AnnotationType: n.AnnotType.String(),
			}
			ch <- pn
		}
	}
	return paths
}

func (n Names) processItem(item *db.Item, pathMap map[string]struct{}) error {
	itemReset(item)
	var maxPath int
	paths := make([][]string, 0, len(pathMap))
	for k := range pathMap {
		clades := strings.Split(k, "|")
		if len(clades) > maxPath {
			maxPath = len(clades)
		}
		paths = append(paths, clades)
		item.PathsTotal++
		switch clades[0] {
		case "Animalia":
			item.AnimaliaNum++
		case "Plantae":
			item.PlantaeNum++
		case "Fungi":
			item.FungiNum++
		case "Bacteria":
			item.BacteriaNum++
		}
	}
	item.MajorKingdom, item.KingdomPercent = getKingdom(paths)
	item.Context = getContext(paths, maxPath)
	n.GormDB.Save(item)
	return nil
}

func getContext(paths [][]string, maxPath int) string {
	var threshold float32 = 0.5
	if maxPath == 0 {
		return ""
	}
	data := make([]map[string]int, maxPath)
	for _, path := range paths {
		for i, v := range path {
			if data[i] == nil {
				data[i] = make(map[string]int)
			}
			data[i][v]++
		}
	}
	var context string
	for _, mp := range data {
		maxTaxon := ""
		var total float32 = 0.0
		var max float32 = 0.0
		for k, v := range mp {
			v := float32(v)
			total += v
			if v > max {
				max = v
				maxTaxon = k
			}
		}
		if max/total < threshold {
			return context
		} else {
			context = maxTaxon
		}
	}
	return context
}

func getKingdom(paths [][]string) (string, uint) {
	if len(paths) == 0 {
		return "", 0
	}
	res := make(map[string]int)
	for _, v := range paths {
		if len(v) < 3 {
			continue
		}
		res[v[0]]++
	}
	var kdm string
	var max, total int
	for k, v := range res {
		total += v
		if v > max {
			max = v
			kdm = k
		}
	}
	percent := (100 * float32(max)) / float32(total)

	if total == 0 {
		return "", 0
	}
	return kdm, uint(percent)
}

func itemReset(item *db.Item) {
	item.PathsTotal = 0
	item.AnimaliaNum = 0
	item.PlantaeNum = 0
	item.FungiNum = 0
	item.BacteriaNum = 0
	item.KingdomPercent = 0
	item.MajorKingdom = ""
	item.Context = ""
}
