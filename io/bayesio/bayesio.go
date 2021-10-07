package bayesio

import (
	_ "embed"

	"github.com/gnames/bayes"
	"github.com/gnames/bhlnames/ent/nlp"
)

//go:embed data/bayes.json
var bayesData []byte

type bayesio struct{}

func New() nlp.NLP {
	return bayesio{}
}

func (b bayesio) Load() bayes.Bayes {
	nb := bayes.New()
	nb.Load(bayesData)
	return nb
}
