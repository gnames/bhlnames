package title_matcher

//go:generate counterfeiter -o titlematchertest/fake_title_matcher.go . TitleMatcher

// TitleMatcher allows to make a match of a journal/book title with a
// biodiversity reference.
type TitleMatcher interface {
	// MatchTitlesBHL takes a reference-string and returns back IDs of matched
	// BHL titles.
	TitlesBHL(refString string) (map[int][]string, error)
}
