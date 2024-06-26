package bayesio

import (
	_ "embed"

	"github.com/gnames/bayes"
	"github.com/gnames/bhlnames/internal/ent/nlp"
)

//go:embed data/bayes.json
var bayesData []byte

type bayesio struct{}

func New() nlp.NLP {
	return bayesio{}
}

func (b bayesio) LoadPretrainedWeights() bayes.Bayes {
	nb := bayes.New()
	nb.Load(bayesData)
	return nb
}
