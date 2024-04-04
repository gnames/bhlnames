// package nlp provides interface to methods for
// natural language processing.
package nlp

import "github.com/gnames/bayes"

// NLP interface provides methods to load NLP models.
// Currently only Naive Bayes model is supported.
type NLP interface {
	// LoadPretrainedWeights creates a new Naive Bayes model using pretrained data.
	// The Bayes object is used to find out if a publication reference
	// corresponds to a found metadta from BHL.
	LoadPretrainedWeights() bayes.Bayes
}
