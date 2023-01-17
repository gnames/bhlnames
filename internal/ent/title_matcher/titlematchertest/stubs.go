package titlematchertest

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gnames/gnfmt"
	"github.com/stretchr/testify/assert"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

type stubTM struct {
	Ref      string           `json:"ref"`
	TitleIds map[int][]string `json:"titleIds"`
}

// Stubs creates data for stubs. To create such data we use
// testdata/stubs_ref_titles.txt file at first to generate all the titles that
// match a specific reference. The original stubs_ref_titles.txt file was
// created by adding the following lines to Calculate function at
// /ent/score/score.go:44 (v0.1.0)
//
//	dump, _ := json.Marshal(titleIDs)
//	fmt.Printf("{ \"ref\": \"%s\", \"titleIds\":", refString)
//	fmt.Println(string(dump) + " }")
//
// and running TestNomenRefs from bhlnames_test.go.
// go test -run Nomen 2>/dev/null |jk > testdata/stubs_ref_titles.json
func Stubs(t *testing.T) map[string]map[int][]string {
	f, err := os.Open(filepath.Join(
		basepath, "..", "..", "..",
		"testdata", "stubs_ref_titles.json"))
	assert.Nil(t, err)

	scnr := bufio.NewScanner(f)
	scnr.Split(bufio.ScanLines)
	var l []string

	for scnr.Scan() {
		l = append(l, scnr.Text())
	}

	f.Close()

	res := make(map[string]map[int][]string)
	var stm stubTM
	for _, v := range l {
		err = new(gnfmt.GNjson).Decode([]byte(v), &stm)
		assert.Nil(t, err)
		res[stm.Ref] = stm.TitleIds
	}
	return res
}
