package bhlnames_test

import (
	"testing"

	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/reffinder/reffindertest"
	"github.com/gnames/bhlnames/internal/ent/title_matcher/titlematchertest"
	"github.com/gnames/bhlnames/internal/io/bayesio"
	"github.com/gnames/bhlnames/internal/io/titlemio"
	bhlnames "github.com/gnames/bhlnames/pkg"
	"github.com/gnames/bhlnames/pkg/config"
	"github.com/gnames/gnparser"
	"github.com/stretchr/testify/assert"
)

func TestNameRefs(t *testing.T) {
	tests := []struct {
		msg, name, current string
		refNum             int
		err                error
	}{
		{
			"acheniumh", "Achenium lusitanicum Skalitzky, 1884",
			"Achenium nigriventris", 15, nil,
		},
	}

	stubs := reffindertest.Stubs(t)

	cfg := config.New()
	gnp := gnparser.New(gnparser.NewConfig())
	rf := new(reffindertest.FakeRefFinder)
	tm, err := titlemio.New(cfg)
	assert.Nil(t, err)

	opts := []bhlnames.Option{
		bhlnames.OptParser(gnp),
		bhlnames.OptRefFinder(rf),
		bhlnames.OptTitleMatcher(tm),
	}

	bn := bhlnames.New(cfg, opts...)
	defer bn.Close()
	for _, v := range tests {
		opt := input.OptNameString(v.name)
		rf.ReferencesBHLReturns(stubs[v.name], nil)
		inp := input.New(gnp, opt)
		res, err := bn.NameRefs(inp)
		assert.Nil(t, err)
		assert.Equal(t, 1, rf.ReferencesBHLCallCount())
		assert.Equal(t, v.current, res.CurrentCanonical)
		assert.Equal(t, v.refNum, res.ReferenceNumber)
	}
}

func TestNomenRefs(t *testing.T) {
	tests := []struct {
		msg, name, ref                                   string
		itemID, score, scoreAnnot, scoreYear, scoreTitle int
		odds                                             float64
	}{
		{
			"Achenium", "Achenium lusitanicum Skalitzky, 1884",
			"Skalitzky, C. Zwei neue europäische Staphylinenarten aus Portugal. Wiener Entomologische Zeitung, 3 (4): 97-99. (1884).",
			43792, 10, 3, 3, 1, 150,
		},
		{
			"Sagenia", "Sagenia longicruris Christ",
			"Christ. In: Bull. Ac. Géogr. Bot. Mans 250. (1906).",
			26220, 6, 3, 3, 0, 0.5,
		},
		{
			"Hamotus", "Hamotus gracilicornis Reitter, 1882",
			"Reitter, E. Neue Pselaphiden und Scydmaeniden aus Brasilien. Deutsche Entomologische Zeitschrift, 26 (1): 129-152, pl. 5. (1882).",
			103947, 10, 3, 3, 1, 20000,
		},
		{
			"Pseudo", "Pseudotrochalus niger Brenske, 1903",
			"",
			42374, 9, 3, 3, 0, 900,
		},
		// {
		// 	"Ortho", "Orthocarpus attenuatus A.Gray",
		// 	"A. Gray. In: Pacif. Rail. Rep. 4: 121. (1857).",
		// 	100867, 8, 3, 2, 0, 30,
		// },
		{
			"Cyathea", "Cyathea aureonitens Christ",
			"Christ. In: Bull. Herb. Boiss. II, 4: 948. (1904).",
			104955, 8, 3, 1, 1, 200,
		},
		// {
		// 	"Licaria", "Licaria simulans C.K.Allen",
		// 	"C. K. Allen. In: Mem. N. Y. Bot. Gard. 10: No. 5, 55. (1964).",
		// 	150908, 8, 3, 3, 2, 10,
		// },
		{
			"Anthoceros", "Anthoceros muscoides Colenso",
			"Colenso W. A further contribution toward making known the botany of New Zealand. Transactions and Proceedings of the New Zealand Institute 16: 325-363. (1884).",
			105242, 11, 3, 2, 3, 8000,
		},
	}

	stubsRF := reffindertest.Stubs(t)
	stubsTM := titlematchertest.Stubs(t)

	cfg := config.New()
	gnp := gnparser.New(gnparser.NewConfig())
	rf := new(reffindertest.FakeRefFinder)
	tmf := new(titlematchertest.FakeTitleMatcher)

	opts := []bhlnames.Option{
		bhlnames.OptParser(gnp),
		bhlnames.OptRefFinder(rf),
		bhlnames.OptTitleMatcher(tmf),
		bhlnames.OptNLP(bayesio.New()),
	}

	bn := bhlnames.New(cfg, opts...)
	defer bn.Close()

	for _, v := range tests {
		opts := []input.Option{
			input.OptNameString(v.name),
			input.OptRefString(v.ref),
		}
		rf.ReferencesBHLReturns(stubsRF[v.name], nil)
		tmf.TitlesBHLReturns(stubsTM[v.ref], nil)
		inp := input.New(gnp, opts...)
		inp.NomenEvent = true
		res, err := bn.NameRefs(inp)
		assert.Nil(t, err)
		assert.True(t, len(res.References) > 0)
		ref := res.References[0]
		assert.Equal(t, v.itemID, ref.ItemID, v.msg)
		assert.Equal(t, v.score, ref.Score.Total, v.msg)
		assert.Equal(t, v.scoreAnnot, ref.Score.Annot, v.msg)
		assert.Equal(t, v.scoreYear, ref.Score.Year, v.msg)
		assert.Equal(t, v.scoreTitle, ref.Score.RefTitle, v.msg)
		assert.Less(t, v.odds, ref.Score.Odds, v.msg)
	}
}
