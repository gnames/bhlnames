package bhlnames_test

import (
	"testing"

	"github.com/gnames/bhlnames"
	"github.com/gnames/bhlnames/config"
	"github.com/gnames/bhlnames/ent/input"
	"github.com/gnames/bhlnames/ent/reffinder/reffindertest"
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
	bn := bhlnames.New(cfg)
	rf := new(reffindertest.FakeRefFinder)
	gnp := gnparser.New(gnparser.NewConfig())

	for _, v := range tests {
		opt := input.OptNameString(v.name)
		rf.ReferencesBHLReturns(stubs[v.name], nil)
		inp := input.New(gnp, opt)
		res, err := bn.NameRefs(rf, inp)
		assert.Nil(t, err)
		assert.Equal(t, rf.ReferencesBHLCallCount(), 1)
		assert.Equal(t, res.CurrentCanonical, v.current)
		assert.Equal(t, res.ReferenceNumber, v.refNum)
	}
}

func TestNomenRefs(t *testing.T) {
	tests := []struct {
		msg, name                    string
		itemID                       int
		score, scoreAnnot, scoreYear int
	}{

		{"Achenium", "Achenium lusitanicum Skalitzky, 1884", 43792, 30, 15, 15},
		{"Sagenia", "Sagenia longicruris Christ, 1906", 26220, 30, 15, 15},
		{"Hamotus", "Hamotus (Hamotus) gracilicornis Reitter, 1882", 209890, 15, 0, 15},
		{"Pseudo", "Pseudotrochalus niger Brenske, 1903", 42374, 30, 15, 15},
		{"Ortho", "Orthocarpus attenuatus A. Gray, 1857", 100867, 26, 15, 11},
		{"Cyathea", "Cyathea aureonitens Christ, 1904", 123193, 15, 15, 0},
		{"Licaria", "Licaria simulans C. K. Allen, 1964", 15466, 4, 0, 4},
		{"Anthoceros", "Anthoceros muscoides Colenso, 1884", 105242, 26, 15, 11},
	}

	stubs := reffindertest.Stubs(t)
	cfg := config.New()
	bn := bhlnames.New(cfg)
	rf := new(reffindertest.FakeRefFinder)
	gnp := gnparser.New(gnparser.NewConfig())

	for _, v := range tests {
		opt := input.OptNameString(v.name)
		rf.ReferencesBHLReturns(stubs[v.name], nil)
		inp := input.New(gnp, opt)
		res, err := bn.NomenRefs(rf, inp)
		assert.Nil(t, err)
		assert.True(t, len(res.References) > 0)
		ref := res.References[0]
		assert.Equal(t, ref.ItemID, v.itemID, v.msg)
		assert.Equal(t, ref.Score.Total, v.score, v.msg)
		assert.Equal(t, ref.Score.Annot, v.scoreAnnot, v.msg)
		assert.Equal(t, ref.Score.Year, v.scoreYear, v.msg)
	}
}
