package repl

import (
	"bytes"
	"encoding/gob"
)

// Messsage sent FROM client
type ClientMsgBody struct {
	ClientId   []byte
	AuthSecret string
	ReplId     []byte
	Head       int
}

func (v *ClientMsgBody) DecodeFrom(b []byte) error {
	var buf bytes.Buffer
	buf.Write(b)
	return gob.NewDecoder(&buf).Decode(v)
}

func (v *ClientMsgBody) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}
