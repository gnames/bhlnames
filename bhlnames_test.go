package bhlnames_test

import (
	"testing"

	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/reffinder/reffindertest"
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
			"Achenium nigriventre", 13, nil,
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
		assert.Equal(t, rf.ReferencesBHLCallCount(), 1)
		assert.Equal(t, res.CurrentCanonical, v.current)
		assert.Equal(t, res.ReferenceNumber, v.refNum)
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
		assert.Equal(t, ref.ItemID, v.itemID, v.msg)
		assert.Equal(t, ref.Score.Total, v.score, v.msg)
		assert.Equal(t, ref.Score.Annot, v.scoreAnnot, v.msg)
		assert.Equal(t, ref.Score.Year, v.scoreYear, v.msg)
		assert.Equal(t, ref.Score.RefTitle, v.scoreTitle, v.msg)
	}
}
