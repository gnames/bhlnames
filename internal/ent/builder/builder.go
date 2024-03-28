package builder

type Builder interface {
	// ResetData removes all downloaded and generated resources, leaving
	// empty databases and directories.
	ResetData() error

	// ImportData downloads remote datasets to the local file system and generates
	// data needed for bhlnames functionality.
	ImportData() error

	// CalculateTxStats calculates taxonomic statistics for each Item.
	CalculateTxStats() error

	// Close closes all resources used by the Builder.
	Close()
}
