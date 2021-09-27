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

func Stubs(t *testing.T) map[string]*namerefs.NameRefs {
	f, err := os.ReadFile(filepath.Join(basepath, "..", "..", "..", "testdata", "stubs_namerefs.gob"))
	assert.Nil(t, err)
	var nrs []*namerefs.NameRefs
	err = gnfmt.GNgob{}.Decode(f, &nrs)
	assert.Nil(t, err)
	res := make(map[string]*namerefs.NameRefs)
	for i := range nrs {
		res[nrs[i].Input.NameString] = nrs[i]
	}
	return res
}
