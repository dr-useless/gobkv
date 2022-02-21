package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

/*
| 0             | 1             | 2             | 3             |
|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|
+---------------+---------------+---------------+---------------+
| < OP        > | < STATUS    > |                               |
| < GOB DATA LEN UINT32                                       > |
  GOB DATA ...
*/

type Message struct {
	Op     byte
	Status byte
	Body   *bytes.Buffer
}

const MSG_HEADER_LEN = 8

// Serialize & write to a given io.Writer
func (m *Message) WriteTo(w io.Writer) (int64, error) {
	header := m.serializeHeader(uint32(m.Body.Len()))

	nh, err := w.Write(header)
	if err != nil {
		return int64(nh), err
	}

	nd, err := m.Body.WriteTo(w)
	return nd + int64(nh), err
}

// Read & deserialize from a given io.Reader
func (m *Message) ReadFrom(r io.Reader) (int64, error) {
	header := make([]byte, MSG_HEADER_LEN)
	nh, err := r.Read(header)
	if err != nil {
		return int64(nh), err
	}

	dataLen, err := m.deserializeHeader(header)
	if err != nil {
		return int64(nh), err
	}

	var nd int64
	if dataLen > 0 {
		g := int(dataLen) - m.Body.Cap()
		if g > 0 {
			m.Body.Grow(g)
		}
		nd, err = m.Body.ReadFrom(r)
	}

	return nd + int64(nh), err
}

func (m *Message) serializeHeader(gobDataLen uint32) []byte {
	b := make([]byte, MSG_HEADER_LEN)

	// OP
	b[0] = m.Op

	// STATUS
	b[1] = m.Status

	// UNUSED [2:4]

	// GOB DATA LEN [4:8]
	l := make([]byte, 4)
	binary.BigEndian.PutUint32(l, gobDataLen)
	copy(b[4:], l)

	return b
}

func (m *Message) deserializeHeader(b []byte) (uint32, error) {
	if len(b) != MSG_HEADER_LEN {
		return 0, errors.New("invalid header: len(b) != 8")
	}
	m.Op = b[0]
	m.Status = b[1]
	return binary.BigEndian.Uint32(b[4:8]), nil
}
