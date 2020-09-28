package bhlnames

import (
	"database/sql"
	"log"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/bhl"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/db"
	"github.com/gnames/bhlnames/domain/entity"
	"github.com/gnames/bhlnames/refs"
	"github.com/gnames/gnames/lib/format"
	"github.com/gnames/gnames/lib/sys"
	"github.com/jinzhu/gorm"
	jsoniter "github.com/json-iterator/go"
	"gitlab.com/gogna/gnparser"
)

type BHLnames struct {
	config.Config
	KV     *badger.DB
	DB     *sql.DB
	GormDB *gorm.DB
}

func NewBHLnames(cnf config.Config) BHLnames {
	bhln := BHLnames{Config: cnf}
	md := bhl.NewMetaData(bhln.Config)
	bhln.KV = db.InitKeyVal(md.PartDir)
	bhln.DB = db.NewDb(cnf.DB)
	bhln.GormDB = db.NewDbGorm(cnf.DB)
	bhln.initDirs()
	return bhln
}

// Init creates all the needed paths
func (bhln BHLnames) initDirs() {
	var err error
	m := bhl.NewMetaData(bhln.Config)
	dirs := []string{m.DownloadDir, m.KeyValDir, m.PartDir}
	for _, dir := range dirs {
		err = sys.MakeDir(dir)
		if err != nil {
			log.Fatalf("Cannot initiate dir '%s': %s.", dir, err)
		}
	}
}

func (bhln BHLnames) Refs(name string) (*entity.Output, error) {
	md := bhl.NewMetaData(bhln.Config)
	kv := bhln.KV
	r := refs.NewRefs(bhln.DB, bhln.GormDB, md, bhln.JobsNum,
		bhln.SortDesc, bhln.Short, bhln.NoSynonyms)
	gnp := gnparser.NewGNparser()
	output := r.Output(gnp, kv, name)
	return output, nil
}

func (bhln BHLnames) RefsStream(chIn <-chan string, chOut chan<- *entity.RefsResult) {
	md := bhl.NewMetaData(bhln.Config)
	kv := bhln.KV
	var wg sync.WaitGroup
	wg.Add(bhln.JobsNum)
	for i := 0; i < bhln.JobsNum; i++ {
		r := refs.NewRefs(bhln.DB, bhln.GormDB, md, bhln.JobsNum,
			bhln.SortDesc, bhln.Short, bhln.NoSynonyms)
		go r.RefsWorker(kv, chIn, chOut, &wg)
	}
	wg.Wait()
	close(chOut)
}

func FormatOutput(output *entity.Output, f format.Format) string {
	var resByte []byte
	var err error
	var res string

	if f == format.PrettyJSON {
		resByte, err = jsoniter.MarshalIndent(output, "", "  ")
	} else {
		resByte, err = jsoniter.Marshal(output)
	}
	if err != nil {
		log.Println(err)
	}
	res = string(resByte)
	res = strings.Replace(res, "\\u0026", "&", -1)
	return res
}
