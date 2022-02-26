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
	// Append split marker using same buffer
	// This will be stripped by the split func
	buf.Write([]byte("+END"))
	return buf.Bytes(), err
}
