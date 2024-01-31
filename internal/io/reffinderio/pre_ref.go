package reffinderio

import (
	"cmp"
	"fmt"
	"log/slog"
	"slices"
	"strconv"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/bhlnames/internal/ent/refbhl"
	"github.com/gnames/bhlnames/internal/io/db"
)

// updateOutput makes sure that every item part and title get only one unique
// name to avoid information overload.
func (rf reffinderio) updateOutput(o *namerefs.NameRefs, raw []*refRow) error {
	kv := rf.kvDB
	o.ReferenceNumber = len(raw)
	partsMap := make(map[string]*preReference)
	itemsMap := make(map[string]*preReference)
	var preRefs []*preReference
	for _, v := range raw {
		partID, err := findPart(kv, v.pageID)
		if err != nil {
			slog.Error("Cannot find part", "pageID", v.pageID, "error", err)
			return err
		}
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
				rf.gormDB.Where("id = ?", partID).First(part)
				partsMap[id] = &preReference{item: v, part: part}
			} else {
				// prefer annotation
				if ref.item.annotation == "NO_ANNOT" &&
					v.annotation != "NO_ANNOT" {
					rf.gormDB.Where("id = ?", partID).First(part)
					partsMap[id] = &preReference{item: v, part: part}
				}
				// prefer a parsed page number
				if ref.item.pageNum.Int64 == 0 && v.pageNum.Int64 > 0 {
					rf.gormDB.Where("id = ?", partID).First(part)
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
	refs := rf.genReferences(preRefs)
	if rf.withSynonyms {
		o.Synonyms = genSynonyms(refs, o.CurrentCanonical)
	}
	if !rf.withShortenedOutput {
		o.References = refs
	}
	return nil
}

func genMapID(id int, name string) string {
	return strconv.Itoa(id) + "-" + name
}

// genSynonyms collects unique name-strings from references and saves all
// of them except the currently accepted name into slice of strings.
func genSynonyms(refs []*refbhl.ReferenceNameBHL, current string) []string {
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

// checks if a page ID is included into any parts. All pageIDs that correspond
// to a particular `part` are saved to key-value store. So if a pageID is not
// found in the store, it means it is not associated with any `parts`. In such case we return 0.
func findPart(kv *badger.DB, pageID int) (int, error) {
	return db.GetValue(kv, strconv.Itoa(pageID))
}

func getURL(pageID int) string {
	if pageID == 0 {
		return ""
	}
	return fmt.Sprintf("https://www.biodiversitylibrary.org/page/%d", pageID)
}

func (l reffinderio) genReferences(prs []*preReference) []*refbhl.ReferenceNameBHL {
	res := make([]*refbhl.ReferenceNameBHL, len(prs))
	for i, v := range prs {
		if v.part == nil {
			v.part = &db.Part{}
		}
		yr, tp := getYearAggr(v)
		res[i] = &refbhl.ReferenceNameBHL{
			NameData: &refbhl.NameData{
				Name:         v.item.name,
				MatchedName:  v.item.matchedCanonical,
				AnnotNomen:   v.item.annotation,
				EditDistance: v.item.editDistance,
			},
			Reference: refbhl.Reference{
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
				Part: &refbhl.Part{
					DOI:   v.part.DOI,
					ID:    int(v.part.ID),
					Pages: getPartPages(v),
					Name:  v.part.Title,
					Year:  int(v.part.Year.Int32),
				},
				ItemStats: refbhl.ItemStats{
					MainKingdom:        v.item.mainKingdom,
					MainKingdomPercent: v.item.mainKingdomPercent,
					UniqNamesNum:       v.item.namesTotal,
					MainTaxon:          v.item.mainTaxon,
				},
			},
		}
	}
	if l.sortDesc {
		slices.SortStableFunc(res, func(a, b *refbhl.ReferenceNameBHL) int {
			return cmp.Compare(b.YearAggr, a.YearAggr)
		})
	} else {
		slices.SortStableFunc(res, func(a, b *refbhl.ReferenceNameBHL) int {
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
