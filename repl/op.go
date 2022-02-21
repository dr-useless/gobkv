package repl

import (
	"encoding/binary"
	"io"
	"log"
)

/*
| 0             | 1             | 2             | 3             |
|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|
+---------------+---------------+---------------+---------------+
| < OP        > | UNUSED        | < KEY LEN UINT16            > |
| < VALUE LEN UINT64                                            |
|                                                             > |
| < KEY EXPIRES UNIX UINT64                                     |
|                                                             > |
| < MODIFIED UNIX UINT64                                        |
|                                                             > |
  KEY ...
  VALUE ...
*/

type ReplOp struct {
	Op       byte
	Expires  uint64
	Modified uint64
	Key      string
	Value    []byte
}

const REP_HEADER_LEN = 28

// Serialize & write to a given io.Writer
func (op *ReplOp) Write(w io.Writer) error {
	header := op.serializeHeader()
	_, err := w.Write(header)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(op.Key))
	if err != nil {
		return err
	}
	_, err = w.Write(op.Value)
	return err
}

// Read & deserialize from a given io.Reader
func (op *ReplOp) Read(r io.Reader) error {
	header := make([]byte, REP_HEADER_LEN)
	_, err := r.Read(header)
	if err != nil {
		return err
	}

	keyLen, valLen := op.deserializeHeader(header)

	if keyLen > 0 {
		keyBytes := make([]byte, keyLen)
		_, err = r.Read(keyBytes)
		if err != nil {
			return err
		}
		op.Key = string(keyBytes)
	}

	if valLen > 0 {
		op.Value = make([]byte, valLen)
		_, err = r.Read(op.Value)
	}

	return err
}

func (op *ReplOp) serializeHeader() []byte {
	b := make([]byte, REP_HEADER_LEN)

	// OP [0]
	b[0] = op.Op

	// UNUSED [1]

	// KEY LEN [2:4]
	keyLen := len(op.Key)
	if keyLen > 0 {
		bKeyLen := make([]byte, 2)
		binary.BigEndian.PutUint16(bKeyLen, uint16(keyLen))
		b[2] = bKeyLen[0]
		b[3] = bKeyLen[1]
	}

	b8 := make([]byte, 8)

	// VAL LEN [4:12]
	valLen := len(op.Value)
	if valLen > 0 {
		binary.BigEndian.PutUint64(b8, uint64(valLen))
		copy(b[4:12], b8)
	}

	// EXPIRES [12:20]
	if op.Expires > 0 {
		binary.BigEndian.PutUint64(b8, op.Expires)
		copy(b[12:20], b8)
	}

	// MODIFIED [20:28]
	if op.Expires > 0 {
		binary.BigEndian.PutUint64(b8, op.Modified)
		copy(b[20:28], b8)
	}

	return b
}

func (op *ReplOp) deserializeHeader(b []byte) (int, int) {
	if len(b) != REP_HEADER_LEN {
		log.Println("invalid header")
		return 0, 0
	}
	op.Op = b[0]
	keyLen := int(binary.BigEndian.Uint16(b[2:4]))
	valLen := int(binary.BigEndian.Uint64(b[4:12]))
	op.Expires = binary.BigEndian.Uint64(b[12:20])
	op.Modified = binary.BigEndian.Uint64(b[20:28])
	return keyLen, valLen
}
