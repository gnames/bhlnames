package namebhl

// NameBHL interface provides methods to collect data from BHLindex.
type NameBHL interface {
	PageFilesToIDs() error
	ImportOccurrences() error
	ImportNames() error
}
