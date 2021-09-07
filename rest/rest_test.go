package rest_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	linkent "github.com/gdower/bhlinker/domain/entity"
	"github.com/gnames/bhlnames/domain/entity"
	"github.com/gnames/gnfmt"
	"github.com/stretchr/testify/assert"
)

const url = "http://:8888/"

func TestNameRefs(t *testing.T) {
	t.Run("finds refs for name-strings", func(t *testing.T) {
		var response []*entity.NameRefs
		enc := gnfmt.GNjson{}
		request := []string{
			"Not name", "Bubo bubo", "Pomatomus",
			"Pardosa moesta", "Plantago major var major",
			"Cytospora ribis mitovirus 2",
			"A-shaped rods", "Alb. alba",
			"Diapria conica",
			"Monohamus galloprovincialis Rüschkamp, 1928",
		}
		req, err := enc.Encode(request)
		assert.Nil(t, err)
		r := bytes.NewReader(req)
		resp, err := http.Post(url+"name_refs", "application/x-binary", r)
		assert.Nil(t, err)
		respBytes, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		enc.Decode(respBytes, &response)

		assert.Equal(t, len(response), 10)

		bad := response[0]
		assert.Equal(t, bad.NameString, "Not name")
		assert.Equal(t, len(bad.References), 0)

		moesta := response[3]
		assert.Equal(t, moesta.NameString, "Pardosa moesta")
		assert.Greater(t, len(moesta.References), 10)
		assert.Equal(t, moesta.CurrentCanonical, "Pardosa moesta")

		gall := response[9]
		assert.Equal(t, gall.NameString, "Monohamus galloprovincialis Rüschkamp, 1928")
		assert.Equal(t, gall.Canonical, "Monohamus galloprovincialis")
		assert.Equal(t, gall.CurrentCanonical, "Monochamus galloprovincialis")
		assert.Greater(t, len(gall.References), 2)
		assert.Less(t, len(gall.References), 10)
	})
}

func TestTaxonRefs(t *testing.T) {
	t.Run("finds references for taxons", func(t *testing.T) {
		var response []*entity.NameRefs
		enc := gnfmt.GNjson{}
		request := []string{
			"Not name", "Bubo bubo", "Pomatomus",
			"Pardosa moesta", "Plantago major var major",
			"Cytospora ribis mitovirus 2",
			"A-shaped rods", "Alb. alba",
			"Diapria conica (Fabricius, 1775)",
			"Monohamus galloprovincialis Rüschkamp, 1928",
		}
		req, err := enc.Encode(request)
		assert.Nil(t, err)
		r := bytes.NewReader(req)
		resp, err := http.Post(url+"taxon_refs", "application/x-binary", r)
		assert.Nil(t, err)
		respBytes, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		enc.Decode(respBytes, &response)

		assert.Equal(t, len(response), 10)

		bad := response[0]
		assert.Equal(t, bad.NameString, "Not name")
		assert.Equal(t, bad.Canonical, "")
		assert.Equal(t, len(bad.References), 0)

		conica := response[8]
		assert.Equal(t, conica.NameString, "Diapria conica (Fabricius, 1775)")
		assert.Equal(t, conica.Canonical, "Diapria conica")
		assert.Equal(t, len(conica.References), 0)

		gall := response[9]
		assert.Equal(t, gall.NameString, "Monohamus galloprovincialis Rüschkamp, 1928")
		assert.Equal(t, gall.Canonical, "Monohamus galloprovincialis")
		assert.Equal(t, gall.CurrentCanonical, "Monochamus galloprovincialis")
		assert.Greater(t, len(gall.References), 100)
		assert.Less(t, len(gall.References), 200)
	})
}

func TestNomenRefs(t *testing.T) {
	t.Run("finds referencers for nomenclatural events", func(t *testing.T) {
		var response []linkent.Output
		enc := gnfmt.GNjson{}
		request := []linkent.Input{
			{
				ID:        "",
				Name:      linkent.Name{Canonical: "Sagenia longicruris"},
				Reference: linkent.Reference{Year: "1906"},
			},
			{
				ID:        "1",
				Name:      linkent.Name{NameString: "Pseudotrochalus niger"},
				Reference: linkent.Reference{Year: "1903"},
			},
			{
				ID:        "myid",
				Name:      linkent.Name{NameString: "Diapria conica"},
				Reference: linkent.Reference{Year: "1775"},
			},
		}
		req, err := enc.Encode(request)
		assert.Nil(t, err)
		r := bytes.NewReader(req)
		resp, err := http.Post(url+"nomen_refs", "application/x-binary", r)
		assert.Nil(t, err)
		respBytes, err := ioutil.ReadAll(resp.Body)
		assert.Nil(t, err)
		enc.Decode(respBytes, &response)

		assert.Equal(t, len(response), 3)

		match := response[0]
		assert.Greater(t, len(match.InputID), 10)
		assert.Equal(t, match.BHLref.AnnotNomen, "SP_NOV")
		assert.Greater(t, match.Score.Overall, float32(0))

		nomatch := response[2]
		assert.Equal(t, nomatch.InputID, "myid")
		assert.Nil(t, nomatch.BHLref)
		assert.Equal(t, nomatch.Score.Overall, float32(0.0))
	})
}
