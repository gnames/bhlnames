package reffinderio

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/bhlnames/ent/refbhl"
	"github.com/gnames/bhlnames/io/db"
)

// updateOutput makes sure that every item part and title get only one unique
// name to avoid information overload.
func (l reffinderio) updateOutput(o *namerefs.NameRefs, raw []*row) {
	kv := l.kvDB
	o.ReferenceNumber = len(raw)
	partsMap := make(map[string]*preReference)
	itemsMap := make(map[string]*preReference)
	var preRefs []*preReference
	for _, v := range raw {
		partID := checkPart(kv, v.pageID)
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
				if ref.item.pageNum == 0 && v.pageNum > 0 {
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
				if ref.item.pageNum == 0 && v.pageNum > 0 {
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
	sort.Strings(res)
	return res
}

// checks if a page ID is included into any parts. All pageIDs that correspond
// to a particular `part` are saved to key-value store. So if a pageID is not
// found in the store, it means it is not associated with any `parts`. In such case we return 0.
func checkPart(kv *badger.DB, pageID int) int {
	return db.GetValue(kv, strconv.Itoa(pageID))
}

func getImagesUrl(name string) string {
	q := url.PathEscape(name)
	url := "https://www.google.com/search?tbm=isch&q=%s"
	return fmt.Sprintf(url, q)
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
			PageNum:            v.item.pageNum,
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
			TitleYearStart:     v.item.titleYearStart,
			TitleYearEnd:       v.item.titleYearEnd,
			ItemYearStart:      v.item.yearStart,
			ItemYearEnd:        v.item.yearEnd,
		}
	}
	if l.sortDesc {
		sort.SliceStable(res, func(i, j int) bool {
			return res[i].YearAggr > res[j].YearAggr
		})
	} else {
		sort.SliceStable(res, func(i, j int) bool {
			return res[i].YearAggr < res[j].YearAggr
		})
	}
	return res
}

func getPartPages(pr *preReference) string {
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
	item = pr.item.yearStart
	title = pr.item.titleYearStart
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
