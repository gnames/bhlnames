package dictio_test

import (
	"testing"

	"github.com/gnames/bhlnames/io/dictio"
	"github.com/stretchr/testify/assert"
)

func TestRemovableWords(t *testing.T) {
	d := dictio.New()
	excl, err := d.ShortWords()
	assert.Nil(t, err)
	assert.True(t, len(excl) > 100)
	_, ok := excl["der"]
	assert.True(t, ok)
}
