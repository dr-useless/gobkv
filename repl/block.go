package repl

import (
	"bytes"
	"encoding/gob"

	"github.com/intob/gobkv/store"
)

// Block for replication
type RBlock map[string]store.Slot

func (v *RBlock) DecodeFrom(b []byte) error {
	var buf bytes.Buffer
	buf.Write(b)
	return gob.NewDecoder(&buf).Decode(v)
}

func (v *RBlock) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}
