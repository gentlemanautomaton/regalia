package regalia

import (
	"errors"
	"io"
)

type fstHeader struct {
}

type header [128]byte

// Rename to unmarshal?

func (h *header) ReadFrom(r io.Reader) (n int64, err error) {
	c, err := r.Read(h[0:1])
	n += int64(c)
	if err != nil {
		return
	}
	if h.Len() > 127 {
		return 1, errors.New("header: length exceeds 128 bytes")
	}
	c, err = r.Read(h[1:h.Len()])
	n += int64(c)
	return
}

func (h *header) Len() int {
	return int(h[0])
}

func (h *header) Offset() int {
	return int(h[1])
}

func encodeFST(s Stream) {

}

func parseFST(s Stream) error {
	r, err := s.Reader()
	if err != nil {
		return err
	}
	var h header
	n, err := h.ReadFrom(r)
	if err != nil && err != io.EOF {
		return err
	}
	_ = n
	//remaining := s.Offset(n).Reader()
	//v := p[h.DataOffset:]
	//v.Header()
	return nil
}

// Cursor is a regalia map cursor that allows retrieval of keys and values.
type Cursor struct {
	Stream
	Key [MaxKey]byte
}

// ReadKey returns the next key from the FST and advances the cursor. If there
// are no more keys nil will be returned.
func (c *Cursor) ReadKey() (key []byte) {
	return nil
}

// ReadKeyValue returns the next key/value pari from the FST and advances the
// cursor.
func (c *Cursor) ReadKeyValue() (key []byte, value []byte) {
	return nil, nil
}
