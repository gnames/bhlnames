package builderio

import (
	"context"
	"database/sql"
	"log/slog"
	"math"

	"github.com/gnames/bhlnames/internal/ent/txstats"
	"github.com/gnames/bhlnames/internal/io/dbio"
	gnstats "github.com/gnames/gnstats/ent/stats"
	"github.com/jackc/pgx/v5"
)

func (b builderio) maxItemID() (int, error) {
	var res int
	err := b.db.QueryRow(
		context.Background(),
		"SELECT max(id) from items",
	).Scan(&res)
	if err != nil {
		slog.Error("Cannot find max item's id", "error", err)
		return res, err
	}
	return res, nil
}

func (b builderio) addStatsToItems(
	ctx context.Context,
	chIn <-chan []txstats.ItemTaxa,
) error {
	columns := []string{
		"id", "names_total", "main_taxon", "main_taxon_rank", "main_taxon_percent",
		"main_kingdom", "main_kingdom_percent", "animalia_num", "plantae_num",
		"fungi_num", "bacteria_num", "main_phylum", "main_phylum_percent",
		"main_class", "main_class_percent", "main_order", "main_order_percent",
		"main_family", "main_family_percent", "main_genus", "main_genus_percent",
	}

	var count int
	for taxa := range chIn {
		rows := make([][]any, 0, len(taxa))
		count += len(taxa)

		var taxon, taxonRank, kingdom, phylum, class, order,
			family, genus sql.NullString
		var taxonPcnt, kingdomPcnt, phylumPcnt, classPcnt, orderPcnt,
			familyPcnt, genusPcnt sql.NullInt16
		var total, animNum, plantNum, fungiNum, bactNum uint
		var st gnstats.Stats

		for _, v := range taxa {
			st = gnstats.New(v.Taxa, 0.5)
			total = uint(st.NamesNum)
			animNum, plantNum, fungiNum, bactNum = kingdomDistribution(st)
			taxon, taxonRank, kingdom, phylum, class, order, family,
				genus = statStrings(st)
			taxonPcnt, kingdomPcnt, phylumPcnt, classPcnt, orderPcnt,
				familyPcnt, genusPcnt = statInts(st)

			row := []any{
				v.ItemID, total, taxon, taxonRank, taxonPcnt,
				kingdom, kingdomPcnt, animNum, plantNum,
				fungiNum, bactNum, phylum, phylumPcnt,
				class, classPcnt, order, orderPcnt,
				family, familyPcnt, genus, genusPcnt,
			}
			rows = append(rows, row)
		}
		_, err := dbio.InsertRows(b.db, "item_stats", columns, rows)
		if err != nil {
			slog.Error("Cannot insert rows to item_stats table", "error", err)
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}

func statInts(st gnstats.Stats) (
	sql.NullInt16, sql.NullInt16, sql.NullInt16,
	sql.NullInt16, sql.NullInt16, sql.NullInt16, sql.NullInt16) {
	var taxonPcnt, kingdomPcnt, phylumPcnt, classPcnt, orderPcnt,
		familyPcnt, genusPcnt sql.NullInt16
	if st.MainTaxon.Name != "" {
		taxonPcnt = floatToNullInt(st.MainTaxonPercentage)
	}
	if st.Kingdom.Name != "" {
		kingdomPcnt = floatToNullInt(st.KingdomPercentage)
	}
	if st.Phylum.Name != "" {
		phylumPcnt = floatToNullInt(st.PhylumPercentage)
	}
	if st.Class.Name != "" {
		classPcnt = floatToNullInt(st.ClassPercentage)
	}
	if st.Order.Name != "" {
		orderPcnt = floatToNullInt(st.OrderPercentage)
	}
	if st.Family.Name != "" {
		familyPcnt = floatToNullInt(st.FamilyPercentage)
	}
	if st.Genus.Name != "" {
		genusPcnt = floatToNullInt(st.GenusPercentage)
	}
	return taxonPcnt, kingdomPcnt, phylumPcnt, classPcnt, orderPcnt, familyPcnt, genusPcnt
}

func floatToNullInt(f float32) sql.NullInt16 {
	i := math.Round(float64(f) * 100)
	return sql.NullInt16{Int16: int16(i), Valid: true}
}

func statStrings(st gnstats.Stats) (
	sql.NullString, sql.NullString,
	sql.NullString, sql.NullString, sql.NullString, sql.NullString,
	sql.NullString, sql.NullString) {
	var taxon, taxonRank, kingdom, phylum, class, order, family, genus sql.NullString
	if st.MainTaxon.Name != "" {
		taxon = sql.NullString{String: st.MainTaxon.Name, Valid: true}
		taxonRank = sql.NullString{String: st.MainTaxon.RankStr, Valid: true}
	}
	if st.Kingdom.Name != "" {
		kingdom = sql.NullString{String: st.Kingdom.Name, Valid: true}
	}
	if st.Phylum.Name != "" {
		phylum = sql.NullString{String: st.Phylum.Name, Valid: true}
	}
	if st.Class.Name != "" {
		class = sql.NullString{String: st.Class.Name, Valid: true}
	}
	if st.Order.Name != "" {
		order = sql.NullString{String: st.Order.Name, Valid: true}
	}
	if st.Family.Name != "" {
		family = sql.NullString{String: st.Family.Name, Valid: true}
	}
	if st.Genus.Name != "" {
		genus = sql.NullString{String: st.Genus.Name, Valid: true}
	}
	return taxon, taxonRank, kingdom, phylum, class, order, family, genus
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
	var rows pgx.Rows
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
	rows, err = b.db.Query(context.Background(), q, id, id+limit)
	if err != nil {
		slog.Error("Cannot get Items data", "error", err)
		return nil, err
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
			slog.Error("Cannot scan Items data", "error", err)
			return nil, err
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
