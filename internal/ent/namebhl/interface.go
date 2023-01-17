package namebhl

import "github.com/bits-and-blooms/bloom/v3"

// NameBHL interface provides methods to collect data from BHLindex.
type NameBHL interface {
	ImportNames() (*bloom.BloomFilter, error)
	ImportOccurrences(*bloom.BloomFilter) error
}
