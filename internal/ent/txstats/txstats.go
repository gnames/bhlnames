package txstats

import (
	"strings"

	gnstats "github.com/gnames/gnstats/ent/stats"
)

type ItemTaxa struct {
	ItemID int
	Taxa   []gnstats.Hierarchy
}

type TxStats struct {
	Classification string
	Ranks          string
	IDs            string
}

func (t TxStats) Taxons() []gnstats.Taxon {
	cls := strings.Split(t.Classification, "|")
	ranks := strings.Split(t.Ranks, "|")
	ids := strings.Split(t.IDs, "|")
	res := make([]gnstats.Taxon, len(cls))

	for i, v := range cls {
		rank := gnstats.NewRank(ranks[i])
		taxon := gnstats.Taxon{
			ID:      ids[i],
			Name:    v,
			RankStr: ranks[i],
			Rank:    rank,
		}
		res[i] = taxon
	}

	return res
}
