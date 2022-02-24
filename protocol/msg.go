package protocol

import (
	"bytes"
	"encoding/gob"
)

// Msg body for normal ops
type Msg struct {
	Op      byte
	Status  byte
	Key     string
	Value   []byte
	Expires int64
	Keys    []string
}

func (v *Msg) DecodeFrom(b []byte) error {
	var buf bytes.Buffer
	buf.Write(b)
	return gob.NewDecoder(&buf).Decode(v)
}

func (v *Msg) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}
