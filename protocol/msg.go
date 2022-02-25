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

func DecodeMsg(b []byte) (*Msg, error) {
	msg := &Msg{}
	var buf bytes.Buffer
	buf.Write(b)
	err := gob.NewDecoder(&buf).Decode(msg)
	return msg, err
}

func EncodeMsg(msg *Msg) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(msg)
	return buf.Bytes(), err
}
