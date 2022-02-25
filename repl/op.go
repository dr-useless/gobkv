package repl

import (
	"bytes"
	"encoding/gob"
)

type Op struct {
	Op       byte
	Expires  int64
	Modified int64
	Key      string
	Value    []byte
}

func (v *Op) DecodeFrom(b []byte) error {
	var buf bytes.Buffer
	buf.Write(b)
	return gob.NewDecoder(&buf).Decode(v)
}

func (v *Op) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}
