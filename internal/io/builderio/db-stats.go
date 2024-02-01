package builderio

import (
	"context"
	"fmt"

	"github.com/gnames/bhlnames/internal/ent/txstats"
	"github.com/gnames/bhlnames/internal/io/db"
	gnstats "github.com/gnames/gnstats/ent/stats"
	"github.com/jackc/pgx/v5"
)

func (b builderio) itemsNum() (int, error) {
	var res int
	ctx := context.Background()
	err := b.DB.QueryRow(ctx, "SELECT count(*) from items").Scan(&res)
	return res, err
}

func (b builderio) addStatsToItems(chIn <-chan []txstats.ItemTaxa) error {
	var err error
	columns := []string{
		"id", "names_total", "main_taxon", "main_taxon_rank", "main_taxon_percent",
		"main_kingdom", "main_kingdom_percent", "animalia_num", "plantae_num",
		"fungi_num", "bacteria_num", "main_phylum", "main_phylum_percent",
		"main_class", "main_class_percent", "main_order", "main_order_percent",
		"main_family", "main_family_percent", "main_genus", "main_genus_percent",
	}

	for taxa := range chIn {
		rows := make([][]any, 0, len(taxa))
		var ist db.ItemStats

		for i, v := range taxa {
			st := gnstats.New(v.Taxa, 0.5)
			ist.NamesTotal = uint(st.NamesNum)
			addKingdomDistribution(&ist, st)
			addStatStrings(&ist, st)
			addStatInts(&ist, st)
			row := []any{
				v.ItemID, ist.NamesTotal, ist.MainTaxon, ist.MainTaxonRank,
				ist.MainTaxonPercent, ist.MainKingdom, ist.MainKingdomPercent,
				ist.AnimaliaNum, ist.PlantaeNum, ist.FungiNum, ist.BacteriaNum,
				ist.MainPhylum, ist.MainPhylumPercent, ist.MainClass,
				ist.MainClassPercent, ist.MainOrder, ist.MainOrderPercent,
				ist.MainFamily, ist.MainFamilyPercent, ist.MainGenus,
				ist.MainGenusPercent,
			}
			rows[i] = row
		}
		_, err = db.InsertRows(b.DB, "items_stats", columns, rows)
		if err != nil {
			return err
		}
	}
	return nil
}

func addStatInts(ist *db.ItemStats, st gnstats.Stats) {
	ist.MainTaxonPercent = uint(st.MainTaxonPercentage)
	ist.MainKingdomPercent = uint(st.KingdomPercentage)
	ist.MainPhylumPercent = uint(st.PhylumPercentage)
	ist.MainClassPercent = uint(st.ClassPercentage)
	ist.MainOrderPercent = uint(st.OrderPercentage)
	ist.MainFamilyPercent = uint(st.FamilyPercentage)
	ist.MainGenusPercent = uint(st.GenusPercentage)
}

func addStatStrings(ist *db.ItemStats, st gnstats.Stats) {
	ist.MainTaxon = st.MainTaxon.Name
	ist.MainTaxonRank = st.MainTaxon.RankStr
	ist.MainKingdom = st.Kingdom.Name
	ist.MainPhylum = st.Phylum.Name
	ist.MainClass = st.Class.Name
	ist.MainOrder = st.Order.Name
	ist.MainFamily = st.Family.Name
	ist.MainGenus = st.Genus.Name
}

func addKingdomDistribution(ist *db.ItemStats, st gnstats.Stats) {
	for _, v := range st.Kingdoms {
		switch v.Name {
		case "Animalia":
			ist.AnimaliaNum = uint(v.NamesNum)
		case "Plantae":
			ist.PlantaeNum = uint(v.NamesNum)
		case "Fungi":
			ist.FungiNum = uint(v.NamesNum)
		case "Bacteria":
			ist.BacteriaNum = uint(v.NamesNum)
		}
	}
}

func (b builderio) getItemsTaxa(id, limit int) ([]txstats.ItemTaxa, error) {
	res := make([]txstats.ItemTaxa, 0, limit)
	var rows pgx.Rows
	var err error

	ctx := context.Background()

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
	rows, err = b.DB.Query(ctx, q, id, id+limit)
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
	if curItemID > 0 {
		res = append(res, txstats.ItemTaxa{ItemID: curItemID, Taxa: hs})
	}

	return res, nil
}
