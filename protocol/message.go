package protocol

import (
	"bufio"
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

// Serialize & write to a given io.Writer
func (m *Message) Write(w io.Writer) {
	buf := bufio.NewWriter(w)

	header := m.serializeHeader()
	buf.Write(header)

	buf.Write([]byte(m.Key))
	buf.Write(m.Value)
	buf.Flush()
}

// Read & deserialize from a given io.Reader
func (m *Message) Read(r io.Reader) {
	buf := bufio.NewReader(r)

	// read first 20 bytes & deserialize header
	header := make([]byte, 20)
	buf.Read(header) // maybe check n

	keyLen, valLen := m.deserializeHeader(header)

	keyBytes := make([]byte, keyLen)
	buf.Read(keyBytes)
	m.Key = string(keyBytes)

	m.Value = make([]byte, valLen)
	buf.Read(m.Value)
}

func (m *Message) serializeHeader() []byte {
	keyLen := len(m.Key)
	valLen := len(m.Value)

	// 20 is total length of fixed fields
	b := make([]byte, 20)

	// OP
	b[0] = m.Op

	// STATUS
	b[1] = m.Status

	// KEY LEN [2:4]
	if keyLen > 0 {
		bKeyLen := make([]byte, 2)
		binary.BigEndian.PutUint16(bKeyLen, uint16(keyLen))
		b[2] = bKeyLen[0]
		b[3] = bKeyLen[1]
	}

	// reuse these
	var offset int
	var i int

	// EXPIRES [4:12]
	if m.Expires > 0 {
		offset = 4
		bExp := make([]byte, 8)
		binary.BigEndian.PutUint64(bExp, m.Expires)
		for i = 0; i < 8; i++ {
			b[offset] = bExp[i]
			offset++
		}
	}

	// VAL LEN [12:20]
	if valLen > 0 {
		offset = 12
		bValLen := make([]byte, 8)
		binary.BigEndian.PutUint64(bValLen, uint64(valLen))
		for i = 0; i < 8; i++ {
			b[offset] = bValLen[i]
			offset++
		}
	}

	return b
}

func (m *Message) deserializeHeader(b []byte) (uint16, uint64) {
	if len(b) != 20 {
		log.Println("invalid header")
		return 0, 0
	}

	m.Op = b[0]

	m.Status = b[1]

	keyLen := binary.BigEndian.Uint16(b[2:4])

	m.Expires = binary.BigEndian.Uint64(b[4:12])

	valLen := binary.BigEndian.Uint64(b[12:20])

	return keyLen, valLen
}
