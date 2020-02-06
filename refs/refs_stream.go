package refs

import (
	"sync"

	"github.com/dgraph-io/badger/v2"
	"gitlab.com/gogna/gnparser"
)

type RefsResult struct {
	Output *Output
	Error  error
}

func (r Refs) RefsWorker(kv *badger.DB, chIn <-chan string,
	chOut chan<- *RefsResult,
	wg *sync.WaitGroup) {
	defer wg.Done()
	defer r.DB.Close()
	for name := range chIn {
		gnp := gnparser.NewGNparser()
		output := r.Output(gnp, kv, name)
		chOut <- &RefsResult{Output: output}
	}
}
