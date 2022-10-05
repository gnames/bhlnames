package builderio

import (
	"github.com/gnames/bhlnames/io/db"
)

func (b builderio) importDataBHL() error {
	var err error
	var titlesMap map[int]*title
	var partDOImap map[int]string
	var itemMap map[uint]string
	var titleDOImap map[int]string

	err = db.Truncate(b.DB, []string{"items", "pages", "parts"})

	if err == nil {
		titleDOImap, partDOImap, err = b.prepareDOI()
	}

	if err == nil {
		titlesMap, err = b.prepareTitle(titleDOImap)
	}

	if err == nil {
		itemMap, err = b.importItem(titlesMap)
	}

	if err == nil {
		err = b.importPart(partDOImap)
	}

	if err == nil {
		err = b.importPage(itemMap)
	}

	if err == nil {
		ts := newTitleStore(b.Config, titlesMap)
		return ts.setup()
	}

	return nil
}
