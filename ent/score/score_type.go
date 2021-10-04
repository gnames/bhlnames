package score

type ScoreType int

const (
	Year ScoreType = iota
	Annot
	RefTitle
	RefVolume
	RefPages
)
