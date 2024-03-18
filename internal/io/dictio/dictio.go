package dictio

import (
	"embed"
	"io"
	"strings"

	"github.com/gnames/bhlnames/internal/ent/dict"
	"github.com/gnames/bhlnames/internal/ent/str"
)

var shortWords map[string]struct{}

//go:embed data
var data embed.FS

type dictIO struct{}

func New() dict.Dict {
	return dictIO{}
}

func (d dictIO) ShortWords() (map[string]struct{}, error) {
	if shortWords != nil {
		return shortWords, nil
	}
	f, err := data.Open("data/excluded.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	txt, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	words := strings.Split(string(txt), "\n")

	res := make(map[string]struct{})
	for i := range words {
		line, err := str.UtfToAscii(words[i])
		if err != nil {
			return nil, err
		}

		res[line] = struct{}{}
	}
	return res, nil
}
