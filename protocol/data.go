package protocol

import (
	"bytes"
	"encoding/gob"
	"errors"
)

// Msg body for normal ops
// Implements Serializable
type Data struct {
	Key     string
	Value   []byte
	Expires int64
	Keys    []string
}

func (d *Data) DecodeFrom(b []byte) error {
	var buf bytes.Buffer
	_, err := buf.Write(b)
	if err != nil {
		return errors.New("failed to decode data: " + err.Error())
	}
	return gob.NewDecoder(&buf).Decode(&d)
}

// Shortcut for DecodeFrom.
// Returns a pointer to a Data{} struct from decoded body
func DecodeData(b []byte) (*Data, error) {
	d := Data{}
	err := d.DecodeFrom(b)
	return &d, err
}

func (d *Data) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&d)
	return buf.Bytes(), err
}
