package builderio

import (
	"database/sql"
	"fmt"
	"math"

	"github.com/gnames/bhlnames/ent/txstats"
	gnstats "github.com/gnames/gnstats/ent/stats"
)

func (b builderio) itemsNum() (int, error) {
	var res int
	err := b.DB.QueryRow("SELECT count(*) from items").Scan(&res)
	return res, err
}

func (b builderio) addStatsToItems(chIn <-chan []txstats.ItemTaxa) error {
	var err error
	var pathsTotal int
	var st gnstats.Stats
	for itxs := range chIn {
		for _, t := range itxs {
			st = gnstats.New(t.Taxa, 0.5)
			pathsTotal = st.NamesNum
			majorKingdom := st.Kingdom.Name
			kingdomPercent := math.Round(float64(st.KingdomPercentage))
			mainTaxon := st.MainTaxon.Name
			anim, plant, fungi, bact := kingdomDistribution(st)
			q := `
UPDATE items 
  SET (paths_total, major_kingdom, kingdom_percent, main_taxon,
			       animalia_num, plantae_num, fungi_num, bacteria_num)
      = (v.paths_total::integer, v.major_kingdom, v.kingdom_percent::integer,
      v.main_taxon, v.animalia_num::integer, v.plantae_num::integer,
      v.fungi_num::integer, v.bacteria_num::integer)
  FROM (
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ) AS v(paths_total, major_kingdom, kingdom_percent,
            main_taxon, animalia_num, plantae_num, fungi_num,
            bacteria_num)
        WHERE id = $9;
`
			_, err = b.DB.Exec(q, pathsTotal, majorKingdom, kingdomPercent, mainTaxon,
				anim, plant, fungi, bact, t.ItemID)
			if err != nil {
				err = fmt.Errorf("addStatsToItems: %w", err)
				return err
			}
		}
	}
	return nil
}

func kingdomDistribution(st gnstats.Stats) (uint, uint, uint, uint) {
	var anim, plant, fungi, bact uint
	for _, v := range st.Kingdoms {
		switch v.Name {
		case "Animalia":
			anim = uint(v.NamesNum)
		case "Plantae":
			plant = uint(v.NamesNum)
		case "Fungi":
			fungi = uint(v.NamesNum)
		case "Bacteria":
			bact = uint(v.NamesNum)
		}
	}
	return anim, plant, fungi, bact
}

func (b builderio) getItemsTaxa(id, limit int) ([]txstats.ItemTaxa, error) {
	res := make([]txstats.ItemTaxa, 0, limit)
	var rows *sql.Rows
	var err error

	q := `
SELECT
  i.id, n.classification, n.classification_ranks, n.classification_ids
  FROM items i
    JOIN pages p on i.id = p.item_id
    JOIN name_occurrences o on p.id = o.page_id
    JOIN name_strings n on n.id = o.name_string_id
  where i.id >= $1 and i.id < $2
GROUP BY i.id, n.classification, n.classification_ranks, n.classification_ids
ORDER BY i.id
`
	rows, err = b.DB.Query(q, id, id+limit)
	if err != nil {
		return nil, fmt.Errorf("statItems: %w", err)
	}
	defer rows.Close()

	var curItemID int
	var hs []gnstats.Hierarchy
	for rows.Next() {
		var itemID int
		var ts txstats.TxStats
		err := rows.Scan(
			&itemID, &ts.Classification, &ts.Ranks, &ts.IDs,
		)
		if err != nil {
			return nil, fmt.Errorf("statItems: %w", err)
		}

		if curItemID == 0 {
			curItemID = itemID
		} else if curItemID != itemID {
			res = append(res, txstats.ItemTaxa{ItemID: curItemID, Taxa: hs})
			curItemID = itemID
			hs = make([]gnstats.Hierarchy, 0)
		}
		hs = append(hs, gnstats.Hierarchy(ts))
	}
	res = append(res, txstats.ItemTaxa{ItemID: curItemID, Taxa: hs})

	return res, nil
}
