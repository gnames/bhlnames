package score

import (
	"strconv"
	"strings"

	"github.com/gnames/bhlnames/ent/refbhl"
)

func getVolumeScore(volume int, ref *refbhl.ReferenceBHL) int {

	volString := strconv.Itoa(volume)
	index := strings.Index(ref.Volume, volString)
	if index != -1 && doesMatch(index, index+len(volString)-1, ref.Volume) {
		return 1
	}

	return 0
}

func doesMatch(start, end int, volume string) bool {
	bs := []byte(volume)
	var matchStart, matchEnd bool

	if start == 0 {
		matchStart = true
	} else {
		beforeByte := bs[start-1]
		if '0' > beforeByte || beforeByte > '9' {
			matchStart = true
		}
	}

	if end == len(volume)-1 {
		matchEnd = true
	} else {
		afterByte := bs[end+1]
		if '0' > afterByte || afterByte > '9' {
			matchEnd = true
		}
	}
	return matchStart && matchEnd
}
