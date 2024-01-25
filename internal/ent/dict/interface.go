// package dict contains tools to access hard-coded dictionaries.
package dict

// Dict provides methods to access hard-coded dictionaries.
type Dict interface {
	// ShortWords returns a map of short words. The map is used to
	// filter out short words from the reference titles.
	// The short words are often excluded from abbreviated titles.
	// For example, the title "Journal of the Linnean Society, Botany"
	// is abbreviated as "J. Linn. Soc., Bot." removing short words
	// "of" and "the".
	ShortWords() (map[string]struct{}, error)
}
