package regalia

import "io"

// Stream is a series of data pages that form a single byte stream.
type Stream interface {
	Reader() (io.ReadSeeker, error)  // TODO: Add io.Closer as well?
	Writer() (io.WriteSeeker, error) // TODO: Add io.Closer as well?
	//Page(i int) Page
	//Range(at uint64, count uint64) []byte
	// TODO: Add hash of all pages?
}
