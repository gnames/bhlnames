package reffndio

import (
	"cmp"
	"fmt"
	"log/slog"
	"slices"
	"strconv"

	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/model"
)

// deduplicateResults makes sure that every item part and title get only one unique
// name to avoid information overload.
func (rf reffndio) deduplicateResults(
	inp input.Input,
	o *bhl.RefsByName,
	raw []*refRec,
) error {

	// partsMap's key is  part.ID + canonical
	partsMap := make(map[string]*preReference)
	// itemsMap's key is item.ID + canonical
	itemsMap := make(map[string]*preReference)

	var preRefs []*preReference
	for _, v := range raw {
		part, err := rf.partByID(v.pageID)
		if err != nil {
			slog.Error("Cannot find part", "pageID", v.pageID, "error", err)
			return err
		}
		if part == nil {
			// find the first name in the Item
			id := genMapID(uint(v.itemID), v.matchedCanonical)
			if ref, ok := itemsMap[id]; !ok {
				itemsMap[id] = &preReference{item: v, part: &model.Part{}}
			} else {
				// prefer annotation
				if ref.item.annotation == "NO_ANNOT" && v.annotation != "NO_ANNOT" {
					itemsMap[id] = &preReference{item: v, part: &model.Part{}}
				}
				// prefer a parsed page number
				if ref.item.pageNum.Int64 == 0 && v.pageNum.Int64 > 0 {
					itemsMap[id] = &preReference{item: v, part: &model.Part{}}
				}
			}
		} else {
			// find the first name in a Part
			id := genMapID(part.ID, v.matchedCanonical)
			if ref, ok := partsMap[id]; !ok {

				partsMap[id] = &preReference{item: v, part: part}
			} else {
				if ref.item.annotation == "NO_ANNOT" &&
					v.annotation != "NO_ANNOT" {
					partsMap[id] = &preReference{item: v, part: part}
				}
				// prefer a parsed page number
				if ref.item.pageNum.Int64 == 0 && v.pageNum.Int64 > 0 {
					partsMap[id] = &preReference{item: v, part: part}
				}
			}
		}
	}
	for _, v := range itemsMap {
		preRefs = append(preRefs, v)
	}
	for _, v := range partsMap {
		preRefs = append(preRefs, v)
	}
	refs := rf.getReferences(preRefs, inp.SortDesc)
	if inp.WithTaxon {
		o.Synonyms = getSynonyms(refs, o.CurrentCanonical)
	}
	if !inp.WithShortenedOutput {
		o.References = refs
	}
	return nil
}

func genMapID(id uint, name string) string {
	return strconv.Itoa(int(id)) + "-" + name
}

// getSynonyms collects unique name-strings from references and saves all
// of them except the currently accepted name into slice of strings.
func getSynonyms(refs []*bhl.ReferenceName, current string) []string {
	syn := make(map[string]struct{})
	for _, v := range refs {
		if v.MatchedName != current {
			syn[v.MatchedName] = struct{}{}
		}
	}
	res := make([]string, 0, len(syn))
	for k := range syn {
		res = append(res, k)
	}
	slices.Sort(res)
	return res
}

func getURL(pageID int) string {
	if pageID == 0 {
		return ""
	}
	return fmt.Sprintf("https://www.biodiversitylibrary.org/page/%d", pageID)
}

func (rf reffndio) getReferences(
	prs []*preReference,
	desc bool,
) []*bhl.ReferenceName {
	res := make([]*bhl.ReferenceName, len(prs))
	for i, v := range prs {
		if v.part == nil {
			v.part = &model.Part{}
		}
		yr, tp := getYearAggr(v)
		res[i] = &bhl.ReferenceName{
			NameData: &bhl.NameData{
				Name:         v.item.name,
				MatchedName:  v.item.matchedCanonical,
				AnnotNomen:   v.item.annotation,
				EditDistance: v.item.editDistance,
			},
			Reference: bhl.Reference{
				YearAggr:       yr,
				YearType:       tp,
				ItemID:         v.item.itemID,
				URL:            getURL(v.item.pageID),
				TitleID:        v.item.titleID,
				TitleName:      v.item.titleName,
				Volume:         v.item.volume,
				TitleDOI:       v.item.titleDOI,
				PageID:         v.item.pageID,
				PageNum:        int(v.item.pageNum.Int64),
				TitleYearStart: int(v.item.titleYearStart.Int32),
				TitleYearEnd:   int(v.item.titleYearEnd.Int32),
				ItemYearStart:  int(v.item.yearStart.Int32),
				ItemYearEnd:    int(v.item.yearEnd.Int32),
				Part: &bhl.Part{
					DOI:   v.part.DOI,
					ID:    int(v.part.ID),
					Pages: getPartPages(v),
					Name:  v.part.Title,
					Year:  int(v.part.Year.Int32),
				},
				ItemStats: bhl.ItemStats{
					MainKingdom:        v.item.mainKingdom,
					MainKingdomPercent: v.item.mainKingdomPercent,
					UniqNamesNum:       v.item.namesTotal,
					MainTaxon:          v.item.mainTaxon,
				},
			},
		}
	}
	slices.SortStableFunc(res, func(a, b *bhl.ReferenceName) int {
		if desc {
			return cmp.Compare(b.YearAggr, a.YearAggr)
		} else {
			return cmp.Compare(a.YearAggr, b.YearAggr)
		}
	})
	return res
}

func getPartPages(pr *preReference) string {
	if pr.part == nil {
		return ""
	}
	start := int(pr.part.PageNumStart.Int32)
	end := int(pr.part.PageNumEnd.Int32)
	if start == 0 {
		return ""
	}
	if end == 0 {
		return fmt.Sprintf("%d-?", start)
	}
	return fmt.Sprintf("%d-%d", start, end)
}

func getYearAggr(pr *preReference) (int, string) {
	var part, item, title int
	if pr.part != nil {
		part = int(pr.part.Year.Int32)
	}
	item = int(pr.item.yearStart.Int32)
	title = int(pr.item.titleYearStart.Int32)
	if part > 0 {
		return part, "Part"
	}

	if title > 0 && item < title {
		return title, "Title"
	}

	if item > 0 {
		return item, "Item"
	}
	return 0, "N/A"
}
