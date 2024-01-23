package reffinderio

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/bhlnames/internal/ent/refbhl"
	"github.com/gnames/bhlnames/internal/io/db"
)

// updateOutput makes sure that every item part and title get only one unique
// name to avoid information overload.
func (l reffinderio) updateOutput(o *namerefs.NameRefs, raw []*refRow) {
	kv := l.kvDB
	o.ReferenceNumber = len(raw)
	partsMap := make(map[string]*preReference)
	itemsMap := make(map[string]*preReference)
	var preRefs []*preReference
	for _, v := range raw {
		partID := findPart(kv, v.pageID)
		if partID == 0 {
			// find the first name in the Item
			id := genMapID(v.itemID, v.matchedCanonical)
			if ref, ok := itemsMap[id]; !ok {
				itemsMap[id] = &preReference{item: v, part: &db.Part{}}
			} else {
				// prefer annotation
				if ref.item.annotation == "NO_ANNOT" && v.annotation != "NO_ANNOT" {
					itemsMap[id] = &preReference{item: v, part: &db.Part{}}
				}
				// prefer a parsed page number
				if ref.item.pageNum.Int64 == 0 && v.pageNum.Int64 > 0 {
					itemsMap[id] = &preReference{item: v, part: &db.Part{}}
				}
			}
		} else {
			// find the first name in a Part
			id := genMapID(partID, v.matchedCanonical)
			part := &db.Part{}
			if ref, ok := partsMap[id]; !ok {
				l.gormDB.Where("id = ?", partID).First(part)
				partsMap[id] = &preReference{item: v, part: part}
			} else {
				// prefer annotation
				if ref.item.annotation == "NO_ANNOT" &&
					v.annotation != "NO_ANNOT" {
					l.gormDB.Where("id = ?", partID).First(part)
					partsMap[id] = &preReference{item: v, part: part}
				}
				// prefer a parsed page number
				if ref.item.pageNum.Int64 == 0 && v.pageNum.Int64 > 0 {
					l.gormDB.Where("id = ?", partID).First(part)
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
	refs := l.genReferences(preRefs)
	if l.withSynonyms {
		o.Synonyms = genSynonyms(refs, o.CurrentCanonical)
	}
	if !l.withShortenedOutput {
		o.References = refs
	}
}

func genMapID(id int, name string) string {
	return strconv.Itoa(id) + "-" + name
}

// genSynonyms collects unique name-strings from references and saves all
// of them except the currently accepted name into slice of strings.
func genSynonyms(refs []*refbhl.ReferenceBHL, current string) []string {
	syn := make(map[string]struct{})
	for _, v := range refs {
		if v.MatchName != current {
			syn[v.MatchName] = struct{}{}
		}
	}
	res := make([]string, 0, len(syn))
	for k := range syn {
		res = append(res, k)
	}
	slices.Sort(res)
	return res
}

// checks if a page ID is included into any parts. All pageIDs that correspond
// to a particular `part` are saved to key-value store. So if a pageID is not
// found in the store, it means it is not associated with any `parts`. In such case we return 0.
func findPart(kv *badger.DB, pageID int) int {
	return db.GetValue(kv, strconv.Itoa(pageID))
}

func getURL(pageID int) string {
	if pageID == 0 {
		return ""
	}
	return fmt.Sprintf("https://www.biodiversitylibrary.org/page/%d", pageID)
}

func (l reffinderio) genReferences(prs []*preReference) []*refbhl.ReferenceBHL {
	res := make([]*refbhl.ReferenceBHL, len(prs))
	for i, v := range prs {
		if v.part == nil {
			v.part = &db.Part{}
		}
		yr, tp := getYearAggr(v)
		res[i] = &refbhl.ReferenceBHL{
			YearAggr:           yr,
			YearType:           tp,
			URL:                getURL(v.item.pageID),
			TitleDOI:           v.item.titleDOI,
			PartDOI:            v.part.DOI,
			Name:               v.item.name,
			MatchName:          v.item.matchedCanonical,
			EditDistance:       v.item.editDistance,
			AnnotNomen:         v.item.annotation,
			PageID:             v.item.pageID,
			PageNum:            int(v.item.pageNum.Int64),
			PartID:             int(v.part.ID),
			ItemID:             v.item.itemID,
			TitleID:            v.item.titleID,
			TitleName:          v.item.titleName,
			Volume:             v.item.volume,
			PartPages:          getPartPages(v),
			PartName:           v.part.Title,
			ItemKingdom:        v.item.mainKingdom,
			ItemKingdomPercent: v.item.mainKingdomPercent,
			StatNamesNum:       v.item.namesTotal,
			ItemMainTaxon:      v.item.mainTaxon,
			TitleYearStart:     int(v.item.titleYearStart.Int32),
			TitleYearEnd:       int(v.item.titleYearEnd.Int32),
			ItemYearStart:      int(v.item.yearStart.Int32),
			ItemYearEnd:        int(v.item.yearEnd.Int32),
		}
	}
	if l.sortDesc {
		slices.SortStableFunc(res, func(a, b *refbhl.ReferenceBHL) int {
			return cmp.Compare(b.YearAggr, a.YearAggr)
		})
	} else {
		slices.SortStableFunc(res, func(a, b *refbhl.ReferenceBHL) int {
			return cmp.Compare(a.YearAggr, b.YearAggr)
		})
	}
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
