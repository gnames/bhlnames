package reffinderio

import "github.com/gnames/bhlnames/internal/ent/refbhl"

// allNomenRefsByExternalID returns all putative nomenclatural references for
// a given external id from a data source.
func (rf reffinderio) allNomenRefsByExternalID(
	dataSource,
	id string,
) ([]*refbhl.Reference, error) {
	return nil, nil
}

// bestNomenRefByExternalID returns the best putative nomenclatural reference
// for a given external id from a data source.
func (rf reffinderio) bestNomenRefByExternalID(
	dataSource,
	id string,
) ([]*refbhl.Reference, error) {
	return nil, nil
}
