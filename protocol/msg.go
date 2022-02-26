package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const ErrMsgLen = "msg does not meet length requirements"
const ErrMsgKeyLen = "msg key len field is greater than remaining msg len"

const MSG_LEN_MIN = 22

// Msg body for normal ops
type Msg struct {
	Op      byte
	Status  byte
	Key     string
	Value   []byte
	Expires int64
}

// Deserializes the given byte slice,
// and returns a pointer to Msg or an error
func DecodeMsg(b []byte) (*Msg, error) {
	msg := &Msg{}

	if len(b) < MSG_LEN_MIN {
		return nil, errors.New(ErrMsgLen)
	}

	msg.Op = b[0]
	msg.Status = b[1]

	msg.Expires = int64(binary.BigEndian.Uint64(b[2:18]))

	keyLen := int(binary.BigEndian.Uint16(b[18:22]))
	keyEnd := 22 + keyLen

	if keyLen > 0 {
		if keyEnd > len(b) {
			return nil, errors.New(ErrMsgKeyLen)
		}
		msg.Key = string(b[22:keyEnd])
	}

	// +END is already stripped by scanner split func
	msg.Value = b[keyEnd:]

	return msg, nil
}

// Serializes the given Msg
func EncodeMsg(msg *Msg) ([]byte, error) {
	var buf bytes.Buffer

	err := buf.WriteByte(msg.Op)
	if err != nil {
		return nil, err
	}
	err = buf.WriteByte(msg.Status)
	if err != nil {
		return nil, err
	}

	// Expires
	expBytes := make([]byte, 16)
	if msg.Expires > 0 {
		binary.BigEndian.PutUint64(expBytes, uint64(msg.Expires))
	}
	_, err = buf.Write(expBytes)
	if err != nil {
		return nil, err
	}

	keyBytes := []byte(msg.Key)

	// Key len
	keyLen := len(keyBytes)
	keyLenBytes := make([]byte, 4)
	if keyLen > 0 {
		binary.BigEndian.PutUint16(keyLenBytes, uint16(keyLen))
	}
	_, err = buf.Write(keyLenBytes)
	if err != nil {
		return nil, err
	}

	// Key
	if keyLen > 0 {
		_, err = buf.Write(keyBytes)
		if err != nil {
			return nil, err
		}
	}

	// Value
	if len(msg.Value) > 0 {
		_, err = buf.Write(msg.Value)
		if err != nil {
			return nil, err
		}
	}

	// Append split marker using same buffer
	// This will be stripped by the split func
	_, err = buf.Write([]byte{'+', 'E', 'N', 'D'})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
