package repl

import (
	"bytes"
	"encoding/gob"
)

type ClientMsg struct {
	Id         []byte
	AuthSecret string
	Head       int
}

func (v *ClientMsg) DecodeFrom(b []byte) error {
	var buf bytes.Buffer
	buf.Write(b)
	return gob.NewDecoder(&buf).Decode(v)
}

func (v *ClientMsg) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}
