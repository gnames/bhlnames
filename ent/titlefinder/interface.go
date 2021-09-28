package titlefinder

// TitleFinder sets methods for
type TitleFinder interface {
	Setup() error
	Search(ref string) (titleIDs map[string]int, err error)
}
