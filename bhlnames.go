package bhlnames

import (
	"log"
	"sync"

	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/domain/entity"
	"github.com/gnames/bhlnames/domain/usecase"
	"github.com/gnames/gnlib/sys"
)

type BHLnames struct {
	config.Config
	usecase.Librarian
	usecase.Builder
}

func NewBHLnames(cfg config.Config) BHLnames {
	bhln := BHLnames{Config: cfg}
	bhln.initDirs()
	return bhln
}

func (bhln BHLnames) Refs(name string, opts ...config.Option) (*entity.NameRefs, error) {
	return bhln.Librarian.ReferencesBHL(name, opts...)
}

func (bhln BHLnames) RefsStream(chIn <-chan string,
	chOut chan<- *entity.NameRefs, opts ...config.Option) {
	var wg sync.WaitGroup
	wg.Add(bhln.JobsNum)

	for i := 0; i < bhln.JobsNum; i++ {
		go bhln.RefsWorker(chIn, chOut, &wg, opts...)
	}
	wg.Wait()
	close(chOut)
}

func (bhln BHLnames) RefsWorker(chIn <-chan string, chOut chan<- *entity.NameRefs,
	wg *sync.WaitGroup, opts ...config.Option) {
	defer wg.Done()
	for name := range chIn {
		nameRefs, err := bhln.ReferencesBHL(name, opts...)
		if err != nil {
			log.Println(err)
		}
		chOut <- nameRefs
	}
}

// Init creates all the needed paths
func (bhln BHLnames) initDirs() {
	var err error
	dirs := []string{bhln.DownloadDir, bhln.KeyValDir, bhln.PartDir}
	for _, dir := range dirs {
		err = sys.MakeDir(dir)
		if err != nil {
			log.Fatalf("Cannot initiate dir '%s': %s.", dir, err)
		}
	}
}

func (bhln BHLnames) Initialize() error {
	if bhln.Config.Rebuild {
		bhln.ResetData()
	}
	return bhln.ImportData()
}
