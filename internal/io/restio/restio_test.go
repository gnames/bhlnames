package restio

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/gnames/bhlnames/internal/ent/bhl"
	"github.com/gnames/bhlnames/internal/ent/input"
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

	inp := input.New(gnparser.NewPool(gnparser.NewConfig(), 5), opts...)

	reqBody, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(err)

	r := bytes.NewReader(reqBody)
	resp, err := http.Post(testURL+"/namerefs", "application/json", r)
	assert.Nil(err)

	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)

	var response bhl.RefsByName
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
		input.OptWithTaxon(true),
	}

	inp := input.New(gnparser.NewPool(gnparser.NewConfig(), 5), opts...)

	reqBody, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(err)

	r := bytes.NewReader(reqBody)
	resp, err := http.Post(testURL+"/namerefs", "application/json", r)
	assert.Nil(err)

	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)

	var response bhl.RefsByName
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

func TestNomenRefs(t *testing.T) {
	assert := assert.New(t)
	opts := []input.Option{
		input.OptNameString("Amphioplus lucidus Koehler, 1922"),
		input.OptRefString(
			"Koehler, R. Ophiurans of the Philippine Seas and adjacent waters. Smithsonian Institution United States National Museum Bulletin. 100(5): 1-486. (1922).",
		),
		input.OptWithNomenEvent(true),
	}
	inp := input.New(gnparser.NewPool(gnparser.NewConfig(), 5), opts...)
	reqBody, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(err)

	r := bytes.NewReader(reqBody)
	resp, err := http.Post(testURL+"/namerefs", "application/json", r)
	assert.Nil(err)

	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)

	var response bhl.RefsByName
	err = enc.Decode(respBytes, &response)
	assert.Nil(err)
	assert.Greater(len(response.References), 0)
	for _, v := range response.References {
		assert.Equal("Amphioplus lucidus", v.MatchedName)
	}
}

func TestRefsByPage(t *testing.T) {
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

	enc := gnfmt.GNjson{}

	for _, v := range tests {
		pageID := strconv.Itoa(v.pageID)
		resp, err := http.Get(testURL + "/references/" + pageID)
		assert.Nil(err)

		bs, err := io.ReadAll(resp.Body)
		var res bhl.Reference
		assert.Nil(err)
		err = enc.Decode(bs, &res)
		assert.Nil(err)
		assert.Equal(v.itemID, res.ItemID)
		if v.partIsNil {
			assert.Nil(res.Part)
		} else {
			assert.NotNil(res.Part)
		}
	}
}

func TestCachedRefs(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg, colID              string
		itemID, pageID, quality int
	}{
		{"moesta", "4DJW3", 17664, 1656908, 5},
	}

	for _, v := range tests {
		resp, err := http.Get(testURL + "/cached_refs/" + v.colID)
		assert.Nil(err)
		assert.Equal(http.StatusOK, resp.StatusCode)
		bs, err := io.ReadAll(resp.Body)
		assert.Nil(err)
		var res bhl.RefsByName
		err = enc.Decode(bs, &res)
		assert.Nil(err)
		if len(res.References) == 0 {
			continue
		}
		ref := res.References[0]
		assert.Equal(v.quality, ref.RefMatchQuality)
		assert.Equal(v.itemID, ref.ItemID)
		assert.Equal(v.pageID, ref.PageID)
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

	for _, v := range tests {
		id := strconv.Itoa(v.itemID)
		resp, err := http.Get(testURL + "/items/" + id)
		assert.Nil(err)
		assert.Equal(http.StatusOK, resp.StatusCode)
		bs, err := io.ReadAll(resp.Body)
		assert.Nil(err)
		var res bhl.Item
		err = enc.Decode(bs, &res)
		assert.Nil(err)
		assert.Equal(v.titleID, res.TitleID)
	}
}

func TestTaxonItems(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg, taxon string
		itemNum    int
		itemID     int
	}{
		{"lep", "Lepidoptera", 2000, 18534},
	}
	for _, v := range tests {
		resp, err := http.Get(testURL + "/taxon_items/" + v.taxon)
		assert.Nil(err)
		assert.Equal(http.StatusOK, resp.StatusCode)
		bs, err := io.ReadAll(resp.Body)
		assert.Nil(err)
		var res []bhl.Item
		err = enc.Decode(bs, &res)
		assert.Nil(err)
		assert.GreaterOrEqual(len(res), v.itemNum)
		itm := res[0]
		assert.Equal(v.itemID, itm.ItemID)
	}

}
