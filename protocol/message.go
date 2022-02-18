package protocol

import (
	"bufio"
	"encoding/binary"
	"errors"
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
func (m *Message) Write(w io.Writer) error {
	buf := bufio.NewWriter(w)
	header := m.serializeHeader()
	_, err := buf.Write(header)
	if err != nil {
		return err
	}
	_, err = buf.Write([]byte(m.Key))
	if err != nil {
		return err
	}
	_, err = buf.Write(m.Value)
	buf.Flush()
	return err
}

// Read & deserialize from a given io.Reader
func (m *Message) Read(r io.Reader) error {
	buf := bufio.NewReader(r)
	_, err := buf.Peek(HEADER_LEN)
	if err != nil {
		return err
	}

	// read first HEADER_LEN bytes & deserialize
	header := make([]byte, HEADER_LEN)
	n, err := buf.Read(header)
	if err != nil {
		return err
	}
	if n != 20 {
		return errors.New("invalid header")
	}

	keyLen, valLen := m.deserializeHeader(header)

	if keyLen < 1 {
		return nil
	}

	keyBytes := make([]byte, keyLen)
	n, err = buf.Read(keyBytes)
	if err != nil {
		return err
	}
	if n != int(keyLen) {
		return errors.New("read key length not as defined in header")
	}
	m.Key = string(keyBytes)

	if valLen > 0 {
		m.Value = make([]byte, valLen)
		n, err = buf.Read(m.Value)
		if err != nil {
			return err
		}
		if n != valLen {
			return errors.New("read value length not as defined in header")
		}
	}

	return nil
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
