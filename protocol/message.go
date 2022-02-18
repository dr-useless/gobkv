package protocol

import (
	"encoding/binary"
	"io"
	"log"
)

/*
| 0             | 1             | 2             | 3             |
|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|
+---------------+---------------+---------------+---------------+
| < OP        > | < STATUS    > | < KEY LEN UINT16            > |
| < KEY EXPIRES UNIX UINT64                                     |
|                                                             > |
| < VALUE LEN UINT64                                            |
|                                                             > |
  KEY ...
  VALUE ...
*/

type Message struct {
	Op      byte
	Status  byte
	Expires uint64
	Key     string
	Value   []byte
}

const HEADER_LEN = 20 // bytes

// Serialize & write to a given io.Writer
// TODO: use buffer
// TODO: handle errors
func (m *Message) Write(w io.Writer) error {
	header := m.serializeHeader()
	w.Write(header)
	w.Write([]byte(m.Key))
	w.Write(m.Value)
	return nil
}

// Read & deserialize from a given io.Reader
func (m *Message) Read(r io.Reader) error {
	// read first HEADER_LEN bytes & deserialize
	header := make([]byte, HEADER_LEN)
	_, err := r.Read(header)
	if err != nil {
		return err
	}

	keyLen, valLen := m.deserializeHeader(header)

	if keyLen > 0 {
		keyBytes := make([]byte, keyLen)
		_, err = r.Read(keyBytes)
		if err != nil {
			return err
		}
		m.Key = string(keyBytes)
	}

	if valLen > 0 {
		m.Value = make([]byte, valLen)
		_, err = r.Read(m.Value)
	}

	return err
}

func (m *Message) serializeHeader() []byte {
	// 20 is total length of fixed fields
	b := make([]byte, 20)

	// OP
	b[0] = m.Op

	// STATUS
	b[1] = m.Status

	// KEY LEN [2:4]
	keyLen := len(m.Key)
	if keyLen > 0 {
		bKeyLen := make([]byte, 2)
		binary.BigEndian.PutUint16(bKeyLen, uint16(keyLen))
		b[2] = bKeyLen[0]
		b[3] = bKeyLen[1]
	}

	b8 := make([]byte, 8)

	// EXPIRES [4:12]
	if m.Expires > 0 {
		binary.BigEndian.PutUint64(b8, m.Expires)
		copy(b[4:12], b8)
	}

	// VAL LEN [12:20]
	valLen := len(m.Value)
	if valLen > 0 {
		binary.BigEndian.PutUint64(b8, uint64(valLen))
		copy(b[12:20], b8)
	}

	return b
}

func (m *Message) deserializeHeader(b []byte) (int, int) {
	if len(b) != 20 {
		log.Println("invalid header")
		return 0, 0
	}
	m.Op = b[0]
	m.Status = b[1]
	keyLen := int(binary.BigEndian.Uint16(b[2:4]))
	m.Expires = binary.BigEndian.Uint64(b[4:12])
	valLen := int(binary.BigEndian.Uint64(b[12:20]))
	return keyLen, valLen
}
