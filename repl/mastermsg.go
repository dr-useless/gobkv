package repl

import (
	"bytes"
	"encoding/gob"
)

// Messsage sent FROM master
type MasterMsgBody struct {
	ReplId   []byte
	Head     int
	MustSync bool // if true, client must fully sync
}

func (v *MasterMsgBody) DecodeFrom(b []byte) error {
	var buf bytes.Buffer
	buf.Write(b)
	return gob.NewDecoder(&buf).Decode(v)
}

func (v *MasterMsgBody) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}
