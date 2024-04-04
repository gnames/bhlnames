package dictio_test

import (
	"testing"

	"github.com/gnames/bhlnames/internal/io/dictio"
	"github.com/stretchr/testify/assert"
)

func TestRemovableWords(t *testing.T) {
	assert := assert.New(t)
	d := dictio.New()
	excl, err := d.ShortWords()
	assert.Nil(err)
	assert.True(len(excl) > 100)
	_, ok := excl["der"]
	assert.True(ok)
}
