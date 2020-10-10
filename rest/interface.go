package rest

import (
	linkent "github.com/gdower/bhlinker/domain/entity"
	"github.com/gnames/bhlnames/domain/entity"
)

type APIProvider interface {
	Port() int
	NameRefs(nameStrings []string) []*entity.NameRefs
	TaxonRefs(nameStrings []string) []*entity.NameRefs
	NomenRefs(inputs []linkent.Input) []linkent.Output
}
