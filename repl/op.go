package repl

import (
	"bytes"
	"encoding/gob"
)

type ROp struct {
	Op       byte
	Key      string
	Value    []byte
	Expires  int64
	Modified int64
}

func (v *ROp) DecodeFrom(b []byte) error {
	var buf bytes.Buffer
	buf.Write(b)
	return gob.NewDecoder(&buf).Decode(v)
}

func (v *ROp) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}
