package score

//
// import (
// 	"strconv"
//
// 	"github.com/gnames/bhlnames/ent/namerefs"
// 	"github.com/gnames/bhlnames/ent/refbhl"
// )
//
// func calc(nameRefs *namerefs.NameRefs) {
// 	bhlRefs := nameRefs.References
// 	year := nameRefs.Input.Reference.RefYear
// 	if year == "" {
// 		year = nameRefs.Input.Name.NameYear
// 	}
// 	bestRef := calcRefScore(bhlRefs, year)
// 	if bestRef != nil {
// 		nameRefs.References = []*refbhl.ReferenceBHL{bestRef}
// 	} else {
// 		nameRefs.References = []*refbhl.ReferenceBHL{bestRef}
// 	}
// }
//
// func calcRefScore(bhlRefs []*refbhl.ReferenceBHL, year string) *refbhl.ReferenceBHL {
// 	// get the best ref matching the year
// 	refYear, scoreYear := matchYear(year, bhlRefs)
// 	// get the best ref matching year and annotation
// 	refAnnot, scoreAnnot, scoreYearComposite := matchAnnot(year, bhlRefs)
// 	// combine all best results
// 	scoreComposite := scoreAnnot + scoreYearComposite
// 	// if combined result 0 nothing is found, return nothing
// 	if scoreYear+scoreComposite == 0 {
// 		return nil
// 	}
//
// 	// first best ref is one for year
// 	refBest := refYear
// 	scoreBest := scoreYear
// 	if refBest == nil {
// 		// if year ref is empty, use annot as best
// 		refBest = refAnnot
// 		scoreBest = scoreComposite
// 	} else if refAnnot != nil {
// 		// if year ref and annot ref are different, and annot is not empty,
// 		// compare their scores and pick the biggest.
// 		if scoreComposite > 0 && refBest.PageID != refAnnot.PageID {
// 			if scoreYear < scoreComposite {
// 				refBest = refAnnot
// 				scoreBest = scoreComposite
// 				scoreYear = scoreYearComposite
// 			}
// 		}
// 	}
// 	if refBest == nil {
// 		return nil
// 	}
// 	score := refbhl.Score{Overall: scoreBest, Annot: scoreAnnot, Year: scoreYear}
// 	refBest.Score = score
// 	return refBest
// }
//
// func matchYear(refYear string, refs []*refbhl.ReferenceBHL) (*refbhl.ReferenceBHL, float32) {
// 	yr, err := strconv.Atoi(refYear)
// 	if err != nil {
// 		yr = 0
// 	}
// 	var refBest *refbhl.ReferenceBHL
// 	var score, scoreBest float32
// 	for _, r := range refs {
// 		score = YearScore(yr, r)
// 		if score > scoreBest {
// 			refBest = r
// 			scoreBest = score
// 		}
// 	}
// 	return refBest, scoreBest
// }
//
// func matchAnnot(refYear string, refs []*refbhl.ReferenceBHL) (*refbhl.ReferenceBHL, float32, float32) {
// 	var refBest *refbhl.ReferenceBHL
// 	var scoreAnnot, score float32
// 	for _, r := range refs {
// 		score = AnnotScore(r)
// 		if score > scoreAnnot {
// 			refBest = r
// 			scoreAnnot = score
// 		}
// 	}
// 	var scoreYear float32 = 0
// 	if scoreAnnot > 0 {
// 		yr, err := strconv.Atoi(refYear)
// 		if err == nil {
// 			scoreYear = YearScore(yr, refBest)
// 		}
// 	}
// 	return refBest, scoreAnnot, scoreYear
// }
