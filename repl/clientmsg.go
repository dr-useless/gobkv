package repl

import (
	"bytes"
	"encoding/gob"
	"errors"
)

// First message sent to the repl server
// The head is sent to verify that the client
// is not too far behind for a partial sync,
// if master tail > client head, a full resync is required
type ClientMsg struct {
	Id         []byte
	AuthSecret string
	Head       int
}

func (msg *ClientMsg) DecodeFrom(b []byte) error {
	var buf bytes.Buffer
	_, err := buf.Write(b)
	if err != nil {
		return errors.New("failed to decode repl client msg: " + err.Error())
	}
	return gob.NewDecoder(&buf).Decode(&msg)
}

// Shortcut for DecodeFrom.
// Returns a pointer to a Data{} struct from decoded body
func DecodeReplClientMsg(b []byte) (*ClientMsg, error) {
	msg := ClientMsg{}
	err := msg.DecodeFrom(b)
	return &msg, err
}

func (msg *ClientMsg) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&msg)
	return buf.Bytes(), err
}
