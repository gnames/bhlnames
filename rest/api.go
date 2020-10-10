package rest

import (
	"sync"

	"github.com/gdower/bhlinker"
	linkent "github.com/gdower/bhlinker/domain/entity"
	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/domain/entity"
	"github.com/google/uuid"
)

type API struct {
	bhln bhlnames.BHLnames
}

func NewAPI(bhln bhlnames.BHLnames) API {
	return API{bhln: bhln}
}

func (api API) Port() int {
	return api.bhln.Config.REST.Port
}

func (api API) NameRefs(nameStrings []string) []*entity.NameRefs {
	opts := []config.Option{config.OptNoSynonyms(true)}
	return api.refs(nameStrings, opts...)
}

func (api API) TaxonRefs(nameStrings []string) []*entity.NameRefs {
	opts := []config.Option{config.OptNoSynonyms(false)}
	return api.refs(nameStrings, opts...)
}

func (api API) NomenRefs(inputs []linkent.Input) []linkent.Output {
	res := make([]linkent.Output, len(inputs))
	lnkr := bhlinker.NewBHLinker(api.bhln, api.bhln.JobsNum)
	chIn := make(chan linkent.Input)
	chOut := make(chan linkent.Output)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for _, inp := range inputs {
			chIn <- inp
		}
		close(chIn)
	}()

	go collectNomens(inputs, res, chOut, &wg)()
	lnkr.GetLinks(chIn, chOut)
	wg.Wait()
	return res
}

func collectNomens(
	input []linkent.Input,
	res []linkent.Output,
	chOut <-chan linkent.Output,
	wg *sync.WaitGroup,
) func() {
	fixIDs(input)
	return func() {
		resMap := make(map[string]linkent.Output)
		defer wg.Done()
		for out := range chOut {
			resMap[out.InputID] = out
		}
		for i, inp := range input {
			res[i] = resMap[inp.ID]
		}
	}
}

func fixIDs(input []linkent.Input) {
	inpMap := make(map[string]struct{})
	for i, inp := range input {
		_, ok := inpMap[inp.ID]
		if inp.ID == "" || ok {
			inp.ID = uuid.New().String()
			input[i] = inp
		} else {
			inpMap[inp.ID] = struct{}{}
		}
	}
}

func feedNames(nameStrings []string, chIn chan<- string) {
	for _, n := range nameStrings {
		chIn <- n
	}
	close(chIn)
}

func (api API) refs(nameStrings []string, opts ...config.Option) []*entity.NameRefs {
	res := make([]*entity.NameRefs, len(nameStrings))
	chIn := make(chan string)
	chOut := make(chan *entity.NameRefs)
	var wg sync.WaitGroup
	wg.Add(1)
	go feedNames(nameStrings, chIn)
	go collectRefs(nameStrings, res, chOut, &wg)()
	api.bhln.RefsStream(chIn, chOut, opts...)
	wg.Wait()
	return res
}

func collectRefs(
	nameStrings []string,
	res []*entity.NameRefs,
	chOut <-chan *entity.NameRefs,
	wg *sync.WaitGroup,
) func() {
	return func() {
		resMap := make(map[string]*entity.NameRefs)
		defer wg.Done()
		for nr := range chOut {
			resMap[nr.NameString] = nr
		}
		for i, n := range nameStrings {
			res[i] = resMap[n]
		}
	}
}
