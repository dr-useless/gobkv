package repl

import (
	"bytes"
	"encoding/gob"
)

type Block struct {
	Slots map[string]Slot
}

type Slot struct {
	Value    []byte
	Expires  int64
	Modified int64
}

func (v *Block) DecodeFrom(b []byte) error {
	var buf bytes.Buffer
	buf.Write(b)
	return gob.NewDecoder(&buf).Decode(v)
}

func (v *Block) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}
