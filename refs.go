package bhlnames

import (
	"log"
	"strings"
	"sync"

	"github.com/gnames/bhlnames/db"
	"github.com/gnames/bhlnames/refs"
	jsoniter "github.com/json-iterator/go"
	"gitlab.com/gogna/gnparser"
)

func (bhln BHLnames) Refs(name string) (*refs.Output, error) {
	kv := db.InitKeyVal(bhln.MetaData.PartDir)
	defer kv.Close()
	r := refs.NewRefs(bhln.DbOpts, bhln.MetaData, bhln.JobsNum,
		bhln.SortDesc, bhln.Short, bhln.NoSynonyms)
	gnp := gnparser.NewGNparser()
	output := r.Output(gnp, kv, name)
	return output, r.DB.Close()
}

func FormatOutput(output *refs.Output, format string) string {
	var resByte []byte
	var err error
	var res string

	if format == "pretty" {
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

func RefsStream(bhln BHLnames, chIn <-chan string,
	chOut chan<- *refs.RefsResult) {
	kv := db.InitKeyVal(bhln.MetaData.PartDir)
	defer kv.Close()
	var wg sync.WaitGroup
	wg.Add(bhln.JobsNum)
	for i := 0; i < bhln.JobsNum; i++ {
		r := refs.NewRefs(bhln.DbOpts, bhln.MetaData, bhln.JobsNum,
			bhln.SortDesc, bhln.Short, bhln.NoSynonyms)
		go r.RefsWorker(kv, chIn, chOut, &wg)
	}
	wg.Wait()
	close(chOut)
}
