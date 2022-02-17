package protocol

import (
	"bufio"
	"encoding/binary"
	"io"
	"time"
)

/*
| 0             | 1             | 2             | 3             |
|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|
+---------------+---------------+---------------+---------------+
| < OP        > | < STATUS    > | < KEY LEN UINT16            > |
| < KEY EXPIRES UNIX INT64                                      |
|                                                             > |
| < VALUE LEN UINT64                                            |
|                                                             > |
  KEY ...
  VALUE ...
*/

type KeyLen uint16
type KeyExpires int64
type ValueLen uint64

type Message struct {
	Op      byte
	Status  byte
	Expires time.Time
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

	// for now, terminate with \r\n.
	//buf.WriteByte('\r')
	//buf.WriteByte('\n')
	//buf.WriteByte('.')
}

// Read & deserialize from a given io.Reader
func (m *Message) Read(r io.Reader) {
	buf := bufio.NewReader(r)

	// read first 20 bytes & deserialize header
	header := make([]byte, 20)
	buf.Read(header) // maybe check n

	msg := Message{}
	keyLen, valLen := msg.deserializeHeader(header)

	keyBytes := make([]byte, keyLen)
	buf.Read(keyBytes)

	valBytes := make([]byte, valLen)
	buf.Read(valBytes)
}

func (m *Message) serializeHeader() []byte {
	keyLen := len(m.Key)
	valLen := len(m.Value)

	// 20 is total length of fixed fields
	b := make([]byte, 20)

	b[0] = m.Op

	b[1] = m.Status

	binary.BigEndian.PutUint16(b[2:3], uint16(keyLen))

	binary.BigEndian.PutUint64(b[4:11], uint64(m.Expires.Unix()))

	binary.BigEndian.PutUint64(b[12:20], uint64(valLen))

	return b
}

func (m *Message) deserializeHeader(b []byte) (uint16, uint64) {
	m.Op = b[0]

	m.Status = b[1]

	keyLen := binary.BigEndian.Uint16(b[2:3])

	exp := binary.BigEndian.Uint64(b[4:11])
	m.Expires = time.Unix(int64(exp), 0)

	valLen := binary.BigEndian.Uint64(b[12:20])

	return keyLen, valLen
}
