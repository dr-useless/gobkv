package protocol

import (
	"bytes"
	"encoding/gob"
)

type Manifest []PartManifest

type PartManifest struct {
	PartId []byte
	Blocks []BlockManifest
}

type BlockManifest struct {
	BlockId []byte
	Hash    []byte
}

func (v *Manifest) DecodeFrom(b []byte) error {
	var buf bytes.Buffer
	buf.Write(b)
	return gob.NewDecoder(&buf).Decode(v)
}

func (v *Manifest) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}
