package namebhl

// NameBHL interface provides methods to collect data from BHLindex.
type NameBHL interface {
	PageFilesToIDs() error
	ImportNames() error
	ImportOccurrences() error
}
