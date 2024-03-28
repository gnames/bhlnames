package acstor

type AhoCorasickStore interface {
	Setup() error
	Open() error
	Get(key string) ([]int, error)
}
