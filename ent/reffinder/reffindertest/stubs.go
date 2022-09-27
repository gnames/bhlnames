package reffindertest

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gnames/bhlnames/ent/namerefs"
	"github.com/gnames/gnfmt"
	"github.com/stretchr/testify/assert"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

// Stubs creates data for stubs. To create such data we use
// testdata/test.csv file at first to generate all the references using
// the command:
// bhlnames name -d '\t' -j 1 testdata/test.csv > tmp.json
func Stubs(t *testing.T) map[string]*namerefs.NameRefs {
	f, err := os.ReadFile(filepath.Join(basepath, "..", "..", "..", "testdata", "stubs_namerefs.json"))
	assert.Nil(t, err)
	var nrs []*namerefs.NameRefs
	err = new(gnfmt.GNjson).Decode(f, &nrs)
	assert.Nil(t, err)
	res := make(map[string]*namerefs.NameRefs)
	for i := range nrs {
		res[nrs[i].Input.NameString] = nrs[i]
	}
	return res
}
