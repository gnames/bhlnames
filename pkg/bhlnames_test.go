package bhlnames_test

import (
	"fmt"
	"testing"

	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/io/bayesio"
	"github.com/gnames/bhlnames/internal/io/reffndio"
	"github.com/gnames/bhlnames/internal/io/ttlmchio"
	bhlnames "github.com/gnames/bhlnames/pkg"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/stretchr/testify/assert"
)

var bnG bhlnames.BHLnames

func Init(t *testing.T) bhlnames.BHLnames {
	if bnG != nil {
		return bnG
	}

	cfg := config.New()
	rf, err := reffndio.New(cfg)
	assert.Nil(t, err)
	tm, err := ttlmchio.New(cfg)
	assert.Nil(t, err)

	opts := []bhlnames.Option{
		bhlnames.OptRefFinder(rf),
		bhlnames.OptTitleMatcher(tm),
		bhlnames.OptNLP(bayesio.New()),
	}

	bnG = bhlnames.New(cfg, opts...)
	return bnG
}

func TestNameRefs(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg, name, current string
		refNum             int
		refsLimit          int
		taxon              bool
		desc               bool
		short              bool
		year               int
	}{
		{
			"taxon", "Achenium lusitanicum Skalitzky, 1884",
			"Achenium nigriventris", 12, 0, true, false, false, 1870,
		},
		{
			"taxon,short", "Achenium lusitanicum Skalitzky, 1884",
			"Achenium nigriventris", 0, 0, true, false, true, 1870,
		},
		{
			"taxon,desc", "Achenium lusitanicum Skalitzky, 1884",
			"Achenium nigriventris", 12, 0, true, true, false, 1988,
		},
		{
			"name", "Achenium lusitanicum Skalitzky, 1884",
			"Achenium nigriventris", 4, 0, false, false, false, 1884,
		},
		{
			"2max,desc,taxon", "Achenium lusitanicum Skalitzky, 1884",
			"Achenium nigriventris", 2, 2, true, true, false, 1988,
		},
	}

	bn := Init(t)

	for _, v := range tests {
		inpOpts := []input.Option{
			input.OptRefsLimit(v.refsLimit),
			input.OptSortDesc(v.desc),
			input.OptWithTaxon(v.taxon),
			input.OptNameString(v.name),
			input.OptWithShortenedOutput(v.short),
		}

		inp := input.New(bn.ParserPool(), inpOpts...)
		res, err := bn.NameRefs(inp)
		assert.Nil(err, v.msg)
		assert.Equal(v.current, res.CurrentCanonical, v.msg)
		assert.Equal(v.refNum, len(res.References), v.msg)
		if v.short {
			assert.Empty(res.References, v.msg)
		} else {
			assert.Equal(v.year, res.References[0].YearAggr, v.msg)
		}
	}
}

func TestNameWithRefs(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg        string
		name       string
		ref        string
		nomen      bool
		itemID     int
		score      int
		scoreAnnot int
		scoreYear  int
		scoreTitle int
		odds       float64
	}{
		{
			"not nomen", "Achenium lusitanicum Skalitzky, 1884",
			"Skalitzky, C. Zwei neue europäische Staphylinenarten aus Portugal. Wiener Entomologische Zeitung, 3 (4): 97-99. (1884).",
			false, 43792, 11, 3, 3, 1, 5_000,
		},
		{
			"nomen", "Achenium lusitanicum Skalitzky, 1884",
			"Skalitzky, C. Zwei neue europäische Staphylinenarten aus Portugal. Wiener Entomologische Zeitung, 3 (4): 97-99. (1884).",
			true, 43792, 11, 3, 3, 1, 25_000,
		},
	}

	bn := Init(t)
	for _, v := range tests {
		inpOpts := []input.Option{
			input.OptNameString(v.name),
			input.OptRefString(v.ref),
			input.OptWithNomenEvent(v.nomen),
		}
		inp := input.New(bn.ParserPool(), inpOpts...)
		res, err := bn.NameRefs(inp)
		assert.Nil(err)
		assert.True(len(res.References) > 0)
		ref := res.References[0]
		assert.Equal(v.itemID, ref.ItemID, v.msg)
		assert.Equal(v.score, ref.Score.Total, v.msg)
		assert.Equal(v.scoreAnnot, ref.Score.Annot, v.msg)
		assert.Equal(v.scoreYear, ref.Score.Year, v.msg)
		assert.Equal(v.scoreTitle, ref.Score.RefTitle, v.msg)
		assert.InDelta(v.odds, ref.Score.Odds, 1000.0, v.msg)
	}
}

func TestNomenRefs(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg        string
		name       string
		ref        string
		itemID     int
		score      int
		scoreAnnot int
		scoreYear  int
		scoreTitle int
		odds       float64
	}{
		{
			"Achenium", "Achenium lusitanicum Skalitzky, 1884",
			"Skalitzky, C. Zwei neue europäische Staphylinenarten aus Portugal. Wiener Entomologische Zeitung, 3 (4): 97-99. (1884).",
			43792, 11, 3, 3, 1, 150,
		},
		{
			"Sagenia", "Sagenia longicruris Christ",
			"Christ. In: Bull. Ac. Géogr. Bot. Mans 250. (1906).",
			26220, 6, 3, 3, 0, 0.5,
		},
		{
			"Hamotus", "Hamotus gracilicornis Reitter, 1882",
			"Reitter, E. Neue Pselaphiden und Scydmaeniden aus Brasilien. Deutsche Entomologische Zeitschrift, 26 (1): 129-152, pl. 5. (1882).",
			103947, 11, 3, 3, 1, 20000,
		},
		{
			"Pseudo", "Pseudotrochalus niger Brenske, 1903",
			"",
			42374, 6, 3, 3, 0, 1,
		},
		{
			"Ortho", "Orthocarpus attenuatus A.Gray",
			"A. Gray. In: Pacif. Rail. Rep. 4: 121. (1857).",
			268849, 8, 3, 2, 0, 30,
		},
		{
			"Cyathea", "Cyathea aureonitens Christ",
			"Christ. In: Bull. Herb. Boiss. II, 4: 948. (1904).",
			104955, 8, 3, 1, 1, 200,
		},
		{
			"Licaria", "Licaria simulans C.K.Allen",
			"C. K. Allen. In: Mem. N. Y. Bot. Gard. 10: No. 5, 55. (1964).",
			150908, 9, 3, 3, 2, 10,
		},
		{
			"Anthoceros", "Anthoceros muscoides Colenso",
			"Colenso W. A further contribution toward making known the botany of New Zealand. Transactions and Proceedings of the New Zealand Institute 16: 325-363. (1884).",
			105242, 11, 3, 2, 3, 8000,
		},
	}
	bn := Init(t)

	for _, v := range tests {
		inpOpts := []input.Option{
			input.OptNameString(v.name),
			input.OptRefString(v.ref),
			input.OptWithNomenEvent(true),
		}
		inp := input.New(bn.ParserPool(), inpOpts...)
		res, err := bn.NameRefs(inp)
		assert.Nil(err)
		assert.True(len(res.References) > 0)
		ref := res.References[0]
		assert.Equal(v.itemID, ref.ItemID, v.msg)
		assert.Equal(v.score, ref.Score.Total, v.msg)
		assert.Equal(v.scoreAnnot, ref.Score.Annot, v.msg)
		assert.Equal(v.scoreYear, ref.Score.Year, v.msg)
		assert.Equal(v.scoreTitle, ref.Score.RefTitle, v.msg)
		assert.Less(v.odds, ref.Score.Odds, v.msg)
	}
}

func TestRefByPageID(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg       string
		pageID    int
		itemID    int
		partIsNil bool
	}{
		{"6059125", 6059125, 29372, true},
		{"1656908", 1656908, 17664, false},
	}

	bn := Init(t)
	for _, v := range tests {
		ref, err := bn.RefByPageID(v.pageID)
		assert.Nil(err)
		assert.Equal(v.itemID, ref.ItemID, v.msg)
		if v.partIsNil {
			assert.Nil(ref.Part)
		} else {
			assert.NotNil(ref.Part)
		}
	}
}

func TestItemStats(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg             string
		itemID, titleID int
	}{
		{"73397", 73397, 29889},
	}

	bn := Init(t)
	for _, v := range tests {
		item, err := bn.ItemStats(v.itemID)
		assert.Nil(err)
		fmt.Printf("ITM: %#v\n", item)
		assert.Equal(v.itemID, item.ItemID, v.msg)
		assert.Equal(v.titleID, item.TitleID, v.msg)
	}
}
