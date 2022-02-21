package protocol

import (
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

type Msg struct {
	Op     byte
	Status byte
	Body   []byte
}

const MSG_HEADER_LEN = 8

// Serialize & write to a given io.Writer
func (m *Msg) WriteTo(w io.Writer) (int64, error) {
	header := m.serializeHeader(uint32(len(m.Body)))

	nh, err := w.Write(header)
	if err != nil {
		return int64(nh), err
	}

	var nd int
	if len(m.Body) > 0 {
		nd, err = w.Write(m.Body)
	}

	return int64(nh + nd), err
}

// Read & deserialize from a given io.Reader
func (m *Msg) ReadFrom(r io.Reader) (int64, error) {
	header := make([]byte, MSG_HEADER_LEN)
	nh, err := r.Read(header)
	if err != nil {
		return int64(nh), err
	}

	dataLen, err := m.deserializeHeader(header)
	if err != nil {
		return int64(nh), err
	}

	var nd int
	if dataLen > 0 {
		m.Body = make([]byte, dataLen)
		nd, err = r.Read(m.Body)
		if err != nil {
			return int64(nh + nd), err
		}
	}

	return int64(nh + nd), err
}

func ReadMsgFrom(r io.Reader) (*Msg, error) {
	m := Msg{}
	_, err := m.ReadFrom(r)
	return &m, err
}

func (m *Msg) serializeHeader(gobDataLen uint32) []byte {
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

func (m *Msg) deserializeHeader(b []byte) (uint32, error) {
	if len(b) != MSG_HEADER_LEN {
		return 0, errors.New("invalid header: len(b) != 8")
	}
	m.Op = b[0]
	m.Status = b[1]
	return binary.BigEndian.Uint32(b[4:8]), nil
}
