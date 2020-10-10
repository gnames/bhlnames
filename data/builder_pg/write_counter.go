package builder_pg

import (
	"github.com/gosuri/uiprogress"
)

// WriteCounter counts the number of bytes written to it. It implements to the io.Writer
// interface and we can pass this into io.TeeReader() which will report progress on each
// write cycle.
type WriteCounter struct {
	total int
	n     int
	bar   *uiprogress.Bar
}

// NewWriteCounter is a constructor for WriteCounter type.
func NewWriteCounter(total int) *WriteCounter {
	bar := uiprogress.AddBar(total)
	bar.PrependCompleted()
	counter := &WriteCounter{total: total, bar: bar}
	return counter
}

// Write writes more bytes to WriteCounter and sets the progress for progress bar.
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.n += n
	return n, wc.bar.Set(wc.n)
}
