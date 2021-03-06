package librarian_pg

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/db"
	"github.com/gnames/bhlnames/domain/entity"
)

// updateOutput makes sure that every item part and title get only one unique
// name to avoid information overload.
func (l LibrarianPG) updateOutput(o *entity.NameRefs, raw []*row) {
	kv := l.KV
	o.ReferenceNumber = len(raw)
	partsMap := make(map[string]*preReference)
	itemsMap := make(map[string]*preReference)
	var preRefs []*preReference
	for _, v := range raw {
		partID := checkPart(kv, v.pageID)
		if partID == 0 {
			id := genMapID(v.itemID, v.matchedCanonical)
			if ref, ok := itemsMap[id]; !ok {
				itemsMap[id] = &preReference{item: v, part: &db.Part{}}
			} else if ref.item.annotation == "NO_ANNOT" && v.annotation != "NO_ANNOT" {
				itemsMap[id] = &preReference{item: v, part: &db.Part{}}
			}
		} else {
			id := genMapID(partID, v.matchedCanonical)
			if ref, ok := partsMap[id]; !ok {
				part := &db.Part{}
				l.GormDB.Where("id = ?", partID).First(part)
				if part != nil {
					partsMap[id] = &preReference{item: v, part: part}
				}
			} else if ref.item.annotation == "NO_ANNOT" &&
				v.annotation != "NO_ANNOT" {
				part := &db.Part{}
				l.GormDB.Where("id = ?", partID).First(part)
				if part != nil {
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
	if !l.NoSynonyms {
		o.Synonyms = genSynonyms(refs, o.CurrentCanonical)
	}
	if !l.Short {
		o.References = refs
	}
}

func genMapID(id int, name string) string {
	return strconv.Itoa(id) + "-" + name
}

// genSynonyms collects unique name-strings from references and saves all
// of them except the currently accepted name into slice of strings.
func genSynonyms(refs []*entity.Reference, current string) []string {
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

func (l LibrarianPG) genReferences(prs []*preReference) []*entity.Reference {
	res := make([]*entity.Reference, len(prs))
	for i, v := range prs {
		yr, tp := getYearAggr(v)
		res[i] = &entity.Reference{
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
			PartID:             int(v.part.ID),
			ItemID:             v.item.itemID,
			TitleID:            v.item.titleID,
			TitleName:          v.item.titleName,
			Volume:             v.item.volume,
			PartPages:          getPartPages(v),
			PartName:           v.part.Title,
			ItemKingdom:        v.item.kingdom,
			ItemKingdomPercent: v.item.kingdomPercent,
			StatNamesNum:       v.item.pathsTotal,
			ItemContext:        v.item.context,
			TitleYearStart:     v.item.titleYearStart,
			TitleYearEnd:       v.item.titleYearEnd,
			ItemYearStart:      v.item.yearStart,
			ItemYearEnd:        v.item.yearEnd,
		}
	}
	if l.SortDesc {
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
