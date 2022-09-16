package bhlnames_test

import (
	"testing"

	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/reffinder/reffindertest"
	"github.com/gnames/bhlnames/io/bayesio"
	"github.com/gnames/bhlnames/io/titlemio"
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
			"Achenium nigriventris", 13, nil,
		},
	}

	stubs := reffindertest.Stubs(t)

	cfg := config.New()
	gnp := gnparser.New(gnparser.NewConfig())
	rf := new(reffindertest.FakeRefFinder)
	tm := titlemio.New(cfg)

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
		msg, name, ref                           string
		itemID                                   int
		score, scoreAnnot, scoreYear, scoreTitle int
	}{

		{
			"Achenium", "Achenium lusitanicum Skalitzky, 1884",
			"Skalitzky, C. Zwei neue europäische Staphylinenarten aus Portugal. Wiener Entomologische Zeitung, 3 (4): 97-99. (1884).",
			43792, 31, 15, 15, 1,
		},
		// {
		// 	"Cyathea", "Cyathea aureonitens",
		// 	"Bulletin de l'Herbier Boissier. ser.2 v.4 1904",
		// 	43792, 31, 15, 15, 1,
		// },
		// {
		// 	"Sagenia", "Sagenia longicruris Christ, 1906",
		// 	"",
		// 	26220, 30, 15, 15, 0,
		// },
		// {
		// 	"Hamotus", "Hamotus (Hamotus) gracilicornis Reitter, 1882",
		// 	"",
		// 	103947, 30, 15, 15, 0,
		// },
		// {
		// 	"Pseudo", "Pseudotrochalus niger Brenske, 1903",
		// 	"",
		// 	42374, 30, 15, 15, 0,
		// },
		// {
		// 	"Ortho", "Orthocarpus attenuatus A. Gray, 1857",
		// 	"",
		// 	91693, 30, 15, 15, 0,
		// },
		// {
		// 	"Cyathea", "Cyathea aureonitens Christ, 1904",
		// 	"",
		// 	104955, 19, 15, 4, 0,
		// },
		// {
		// 	"Licaria", "Licaria simulans C. K. Allen, 1964",
		// 	"",
		// 	15466, 19, 15, 4, 0,
		// },
		// {
		// 	"Anthoceros", "Anthoceros muscoides Colenso, 1884",
		// 	"",
		// 	105242, 26, 15, 11, 0,
		// },
	}

	stubs := reffindertest.Stubs(t)

	cfg := config.New()
	gnp := gnparser.New(gnparser.NewConfig())
	rf := new(reffindertest.FakeRefFinder)
	tm := titlemio.New(cfg)

	opts := []bhlnames.Option{
		bhlnames.OptParser(gnp),
		bhlnames.OptRefFinder(rf),
		bhlnames.OptTitleMatcher(tm),
		bhlnames.OptNLP(bayesio.New()),
	}

	bn := bhlnames.New(cfg, opts...)
	defer bn.Close()

	for _, v := range tests {
		opts := []input.Option{
			input.OptNameString(v.name),
			input.OptRefString(v.ref),
		}
		rf.ReferencesBHLReturns(stubs[v.name], nil)
		inp := input.New(gnp, opts...)
		res, err := bn.NomenRefs(inp)
		assert.Nil(t, err)
		assert.True(t, len(res.References) > 0)
		ref := res.References[0]
		assert.Equal(t, v.itemID, ref.ItemID, v.msg)
		assert.Equal(t, v.score, ref.Score.Total, v.msg)
		assert.Equal(t, v.scoreAnnot, ref.Score.Annot, v.msg)
		assert.Equal(t, v.scoreYear, ref.Score.Year, v.msg)
		assert.Equal(t, v.scoreTitle, v.scoreTitle, v.msg)
	}
}
