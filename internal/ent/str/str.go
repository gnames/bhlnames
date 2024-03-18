package str

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Reverse changes order of runes in a string to opposite.
func Reverse(s string) string {
	size := len(s)
	buf := make([]byte, size)
	for start := 0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}
	return string(buf)
}

// UtfToAscii converts all UTF8 runes into closest related ASCII characters.
func UtfToAscii(s string) (string, error) {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, err := transform.String(t, s)
	return s, err
}

// SplitAny splits a string `s` by any rune in `seps` string.
func SplitAny(s string, seps string) []string {
	splitter := func(r rune) bool {
		return strings.ContainsRune(seps, r)
	}
	return strings.FieldsFunc(s, splitter)
}

// IsASCII checks if a string contains only ASCII characters.
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
