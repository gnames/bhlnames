package nlp

import "github.com/gnames/bayes"

type NLP interface {
	Load() bayes.Bayes
}
