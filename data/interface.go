package data

import (
	"github.com/gnames/bhlnames/domain/entity"
	"gitlab.com/gogna/gnparser"
)

type Referencer interface {
	RefsForNameInBHL(name_string string, gnp gnparser.GNparser) (*entity.Output, error)
	Close()
}
