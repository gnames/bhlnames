package acstor

type AhoCorasickStore interface {
	Setup() error
	Get(abbr string) ([]int, error)
}
