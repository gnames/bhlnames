package builderio

import (
	"github.com/gnames/bhlnames/internal/ent/model"
	"github.com/gnames/bhlnames/internal/io/acstorio"
	"github.com/gnames/bhlnames/internal/io/dbio"
)

func (b builderio) importDataBHL() error {
	var err error
	// titleMap has titleID as a key, and Title data as a value.
	var titlesMap map[int]*model.Title
	var partDOImap map[int]string
	var titleDOImap map[int]string

	err = dbio.Truncate(b.db, []string{"items", "pages", "parts", "page_parts"})
	if err != nil {
		return err
	}

	// DOI can belong to either a title or a part, so we need to prepare two
	// lookup maps for the next steps: one for titles and one for parts.
	titleDOImap, partDOImap, err = b.prepareDOI()
	if err != nil {
		return err
	}

	// title data is needed for items, so we prepare it first.
	titlesMap, err = b.prepareTitle(titleDOImap)
	if err != nil {
		return err
	}

	err = b.importItem(titlesMap)
	if err != nil {
		return err
	}

	err = b.importPart(partDOImap)
	if err != nil {
		return err
	}

	err = b.importPage()
	if err != nil {
		return err
	}

	err = b.assignPartsToPages()
	if err != nil {
		return err
	}

	ac, err := acstorio.New(b.cfg, titlesMap)
	if err != nil {
		return err
	}

	// Create AhoCorasickStore where abbreviated titles point to title IDs.
	// It also creates a file with all found abbreviations, that is used
	// lately to get Aho-Coarsick trie.
	err = ac.Setup()
	if err != nil {
		return err
	}

	return nil
}
