package ttlmch

// TitleMatcher allows to make a match of a journal/book title with a
// biodiversity reference.
type TitleMatcher interface {
	// MatchTitlesBHL takes a reference-string and returns back IDs of matched
	// BHL titles.
	TitlesBHL(refString string) (map[int][]string, error)

	// Close cleans database connection.
	Close()
}
