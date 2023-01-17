package dict

type Dict interface {
	ShortWords() (map[string]struct{}, error)
}
