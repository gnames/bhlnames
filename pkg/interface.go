package bhlnames

// BHLnames provides methods for finding references for scientific names
// in the Biodiversity Heritage Library (BHL).
type BHLnames interface {
	// Initialize downloads data about BHL corpus and names found in the
	// corpus. It then imports the data into its storage.
	Initialize() error

	// Close terminates connections to databases and key-value stores.
	Close() error
}
