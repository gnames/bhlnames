package restio

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/gnames/bhlnames/internal/ent/input"
	"github.com/gnames/bhlnames/internal/ent/namerefs"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnparser"
	"github.com/stretchr/testify/assert"
)

var (
	testURL = "http://0.0.0.0:8888/api/v1"
	enc     = gnfmt.GNjson{}
)

func TestPing(t *testing.T) {
	resp, err := http.Get(testURL + "/ping")
	assert.Nil(t, err)

	res, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Equal(t, "pong", string(res))
}

func TestVer(t *testing.T) {
	resp, err := http.Get(testURL + "/version")
	assert.Nil(t, err)

	res, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Contains(t, string(res), "version")
}

func TestNameRefs(t *testing.T) {
	assert := assert.New(t)
	opts := []input.Option{
		input.OptNameString("Achenium lusitanicum Skalitzky, 1884"),
	}

	gnp := gnparser.New(gnparser.NewConfig())
	inp := input.New(gnp, opts...)

	reqBody, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(err)

	r := bytes.NewReader(reqBody)
	resp, err := http.Post(testURL+"/name_refs", "application/json", r)
	assert.Nil(err)

	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)

	var response namerefs.NameRefs
	err = enc.Decode(respBytes, &response)
	assert.Nil(err)
	assert.Greater(response.ReferenceNumber, 0)
	assert.Greater(len(response.References), 0)
	for _, v := range response.References {
		assert.Equal("Achenium lusitanicum", v.MatchedName)
	}
}

func TestTaxonRefs(t *testing.T) {
	assert := assert.New(t)
	opts := []input.Option{
		input.OptNameString("Achenium lusitanicum Skalitzky, 1884"),
	}

	gnp := gnparser.New(gnparser.NewConfig())
	inp := input.New(gnp, opts...)

	reqBody, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(err)

	r := bytes.NewReader(reqBody)
	resp, err := http.Post(testURL+"/taxon_refs", "application/json", r)
	assert.Nil(err)

	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)

	var response namerefs.NameRefs
	err = enc.Decode(respBytes, &response)
	assert.Nil(err)
	assert.Greater(response.ReferenceNumber, 0)
	assert.Greater(len(response.References), 0)
	matchedNames := make(map[string]struct{})
	for _, v := range response.References {
		matchedNames[v.MatchedName] = struct{}{}
	}
	assert.Greater(len(matchedNames), 1)
}

// func TestNomenRefs(t *testing.T) {
// 	assert := assert.New(t)
// 	opts := []input.Option{
// 		input.OptNameString("Amphioplus lucidus Koehler, 1922"),
// 		input.OptRefString(
// 			"Koehler, R. Ophiurans of the Philippine Seas and adjacent waters. Smithsonian Institution United States National Museum Bulletin. 100(5): 1-486. (1922).",
// 		),
// 	}
//
// 	gnp := gnparser.New(gnparser.NewConfig())
// 	inp := input.New(gnp, opts...)
//
// 	reqBody, err := gnfmt.GNjson{}.Encode(inp)
// 	assert.Nil(err)
//
// 	r := bytes.NewReader(reqBody)
// 	resp, err := http.Post(testURL+"/nomen_refs", "application/json", r)
// 	assert.Nil(err)
//
// 	respBytes, err := io.ReadAll(resp.Body)
// 	assert.Nil(err)
//
// 	var response namerefs.NameRefs
// 	err = enc.Decode(respBytes, &response)
// 	assert.Nil(err)
// 	assert.Greater(response.ReferenceNumber, 10)
// }
