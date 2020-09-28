package refs

import (
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/domain/entity"
	"gitlab.com/gogna/gnparser"
)

func (r Refs) RefsWorker(kv *badger.DB, chIn <-chan string,
	chOut chan<- *entity.RefsResult,
	wg *sync.WaitGroup) {
	defer wg.Done()
	defer r.DB.Close()
	for name := range chIn {
		gnp := gnparser.NewGNparser()
		output := r.Output(gnp, kv, name)
		chOut <- &entity.RefsResult{Output: output}
	}
}
