package refs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strconv"

	"github.com/dgraph-io/badger/v2"
	"github.com/gnames/bhlnames/bhl"
	"github.com/gnames/bhlnames/db"
	"github.com/jinzhu/gorm"
	"gitlab.com/gogna/gnparser"
)

type Refs struct {
	PartDir    string
	JobsNum    int
	SortDesc   bool
	Short      bool
	NoSynonyms bool
	DB         *sql.DB
	GormDB     *gorm.DB
}

func NewRefs(dbOpts db.DbOpts, md bhl.MetaData, jobs int, sortDesc,
	short, noSynonyms bool) Refs {
	res := Refs{
		PartDir:    md.PartDir,
		JobsNum:    jobs,
		SortDesc:   sortDesc,
		Short:      short,
		NoSynonyms: noSynonyms,
	}
	res.DB = dbOpts.NewDb()
	res.GormDB = dbOpts.NewDbGorm()
	return res
}

type Output struct {
	NameString       string       `json:"name_string"`
	Canonical        string       `json:"canonical,omitempty"`
	CurrentCanonical string       `json:"current_canonical,omitempty"`
	Synonyms         []string     `json:"synonyms,omitempty"`
	ImagesUrl        string       `json:"images_url,omitempty"`
	ReferenceNumber  int          `json:"refs_num"`
	References       []*Reference `json:"references,omitempty"`
}

type PreReference struct {
	item *Row
	part *db.Part
}

type Reference struct {
	YearAggr           int    `json:"year_aggr"`
	YearType           string `json:"year_type"`
	URL                string `json:"url,omitempty"`
	TitleDOI           string `json:"doi_title,omitempty"`
	PartDOI            string `json:"doi_part,omitempty"`
	Name               string `json:"name"`
	MatchName          string `json:"match_name"`
	EditDistance       int    `json:"edit_distance,omitempty"`
	AnnotNomen         string `json:"annot_nomen,omitempty"`
	PageID             int    `json:"page_id"`
	ItemID             int    `json:"item_id"`
	TitleID            int    `json:"title_id"`
	PartID             int    `json:"part_id,omitempty"`
	TitleName          string `json:"title_name"`
	Volume             string `json:"volume,omitempty"`
	PartPages          string `json:"part_pages,omitempty"`
	PartName           string `json:"part_name,omitempty"`
	ItemKingdom        string `json:"item_kingdom"`
	ItemKingdomPercent int    `json:"item_kingdom_percent"`
	StatNamesNum       int    `json:"stat_names_num"`
	ItemContext        string `json:"item_context"`
	TitleYearStart     int    `json:"title_year_start"`
	TitleYearEnd       int    `json:"title_year_end,omitempty"`
	ItemYearStart      int    `json:"item_year_start,omitempty"`
	ItemYearEnd        int    `json:"item_year_end,omitempty"`
}

type Row struct {
	itemID           int
	titleID          int
	pageID           int
	titleDOI         string
	titleYearStart   int
	titleYearEnd     int
	yearStart        int
	yearEnd          int
	volume           string
	titleName        string
	context          string
	kingdom          string
	kingdomPercent   int
	pathsTotal       int
	nameID           string
	name             string
	annotation       string
	matchedCanonical string
	matchType        string
	editDistance     int
}

// Output takes a name-string and returns back back references where this
// name-string was detected.  If a name string has a nomenclatural annotaion
// (like 'sp. nov.') attached to it somewhere in the reference, we return that
// position, if we do not find such annotation, we return first detection of
// the name-string in the reference.  When a reference is an `item` with
// `parts`, we return one occurrence for every `part`, but also first occurence
// of a name-string in the `item`, if it exists outside of all `parts.
func (r Refs) Output(gnp gnparser.GNparser, kv *badger.DB,
	name string) *Output {
	res := &Output{NameString: name, Canonical: "", CurrentCanonical: "",
		ImagesUrl: "", References: make([]*Reference, 0)}
	can, err := getCanonical(gnp, name)
	if err != nil {
		return res
	}
	res.Canonical = can
	if !r.NoSynonyms {
		res.CurrentCanonical = can
	}
	raw := r.nameQuery(can, "current_canonical")
	if len(raw) == 0 {
		raw = r.matchQuery(res, can)
	}
	if !r.NoSynonyms {
		res.ImagesUrl = getImagesUrl(res.CurrentCanonical)
	}
	r.updateOutput(kv, res, raw)
	return res
}

func getCanonical(gnp gnparser.GNparser, name string) (string, error) {
	ps := gnp.ParseToObject(name)
	if !ps.Parsed {
		return "", errors.New("Could not parse")
	}
	can := ps.Canonical.GetFull()
	return can, nil
}

func (r Refs) nameQuery(name string, field string) []*Row {
	var res []*Row
	var itemID, titleID, pageID int
	var yearStart, yearEnd, titleYearStart, titleYearEnd,
		kingdomPercent, pathsTotal, editDistance sql.NullInt32
	var nameID string
	var titleName, context, majorKingdom, nameString, matchedCanonical,
		matchType, vol, titleDOI, annot sql.NullString
	qs := `SELECT
	itm.id, itm.title_id, pns.page_id, pns.annotation_type, itm.title_year_start,
	itm.title_year_end, itm.year_start, itm.year_end, itm.title_name, itm.vol,
	itm.title_doi, itm.context, itm.major_kingdom, itm.kingdom_percent,
	itm.paths_total, ns.id, ns.name, ns.matched_canonical, ns.match_type,
	ns.edit_distance
	FROM name_strings ns
			JOIN page_name_strings pns ON ns.id = pns.name_string_id
			JOIN pages pg ON pg.id = pns.page_id
			JOIN items itm ON itm.id = pg.item_id
	WHERE ns.%s = '%s'
	ORDER BY title_year_start`
	q := fmt.Sprintf(qs, field, name)

	rows := db.RunQuery(r.DB, q)
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&itemID, &titleID, &pageID, &annot, &titleYearStart,
			&titleYearEnd, &yearStart, &yearEnd, &titleName, &vol, &titleDOI,
			&context, &majorKingdom, &kingdomPercent, &pathsTotal, &nameID,
			&nameString, &matchedCanonical, &matchType, &editDistance)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, &Row{
			itemID:           itemID,
			titleID:          titleID,
			pageID:           pageID,
			titleDOI:         titleDOI.String,
			titleYearStart:   int(titleYearStart.Int32),
			titleYearEnd:     int(titleYearEnd.Int32),
			yearStart:        int(yearStart.Int32),
			yearEnd:          int(yearEnd.Int32),
			titleName:        titleName.String,
			volume:           vol.String,
			context:          context.String,
			kingdom:          majorKingdom.String,
			kingdomPercent:   int(kingdomPercent.Int32),
			pathsTotal:       int(pathsTotal.Int32),
			nameID:           nameID,
			name:             nameString.String,
			annotation:       annot.String,
			matchedCanonical: matchedCanonical.String,
			matchType:        matchType.String,
			editDistance:     int(editDistance.Int32),
		})
	}
	return res
}

func (r Refs) matchQuery(o *Output, name string) []*Row {
	if r.NoSynonyms {
		return r.nameQuery(name, "matched_canonical")
	} else {
		rec := &db.NameString{}
		r.GormDB.Where("matched_canonical = ?", name).First(rec)
		if rec.ID == "" {
			var emptyRes []*Row
			return emptyRes
		}
		o.CurrentCanonical = rec.CurrentCanonical
		return r.nameQuery(o.CurrentCanonical, "current_canonical")
	}
}

func genMapID(id int, name string) string {
	return strconv.Itoa(id) + "-" + name
}

// updateOutput makes sure that every item part and title get only one unique
// name to avoid information overload.
func (r Refs) updateOutput(kv *badger.DB, o *Output, raw []*Row) {
	o.ReferenceNumber = len(raw)
	partsMap := make(map[string]*PreReference)
	itemsMap := make(map[string]*PreReference)
	var preRefs []*PreReference
	for _, v := range raw {
		partID := checkPart(kv, v.pageID)
		if partID == 0 {
			id := genMapID(v.itemID, v.matchedCanonical)
			if ref, ok := itemsMap[id]; !ok {
				itemsMap[id] = &PreReference{item: v, part: &db.Part{}}
			} else if ref.item.annotation == "NO_ANNOT" && v.annotation != "NO_ANNOT" {
				itemsMap[id] = &PreReference{item: v, part: &db.Part{}}
			}
		} else {
			id := genMapID(partID, v.matchedCanonical)
			if ref, ok := partsMap[id]; !ok {
				part := &db.Part{}
				r.GormDB.Where("id = ?", partID).First(part)
				if part != nil {
					partsMap[id] = &PreReference{item: v, part: part}
				}
			} else if ref.item.annotation == "NO_ANNOT" &&
				v.annotation != "NO_ANNOT" {
				part := &db.Part{}
				r.GormDB.Where("id = ?", partID).First(part)
				if part != nil {
					partsMap[id] = &PreReference{item: v, part: part}
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
	refs := r.genReferences(preRefs)
	if !r.NoSynonyms {
		o.Synonyms = genSynonyms(refs, o.CurrentCanonical)
	}
	if !r.Short {
		o.References = refs
	}
}

// genSynonyms collects unique name-strings from references and saves all
// of them except the currently accepted name into slice of strings.
func genSynonyms(refs []*Reference, current string) []string {
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

func (r Refs) genReferences(prs []*PreReference) []*Reference {
	res := make([]*Reference, len(prs))
	for i, v := range prs {
		yr, tp := getYearAggr(v)
		res[i] = &Reference{
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
	if r.SortDesc {
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

func getPartPages(pr *PreReference) string {
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

func getYearAggr(pr *PreReference) (int, string) {
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
