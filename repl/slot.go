package repl

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const ErrMsgLen = "msg does not meet length requirements"
const ErrMsgKeyLen = "msg key len field is greater than remaining msg len"

const REPL_MSG_LEN_MIN = 36

// Slot msg for replication
type ReplMsg struct {
	Expires  int64
	Modified int64
	Key      string
	Value    []byte
}

// Deserializes the given byte slice,
// and returns a pointer to Msg or an error
func DecodeReplMsg(b []byte) (*ReplMsg, error) {
	msg := &ReplMsg{}

	if len(b) < REPL_MSG_LEN_MIN {
		return nil, errors.New(ErrMsgLen)
	}

	msg.Expires = int64(binary.BigEndian.Uint64(b[0:16]))

	msg.Modified = int64(binary.BigEndian.Uint64(b[16:32]))

	keyLen := int(binary.BigEndian.Uint16(b[32:REPL_MSG_LEN_MIN]))
	keyEnd := REPL_MSG_LEN_MIN + keyLen

	if keyLen > 0 {
		if keyEnd > len(b) {
			return nil, errors.New(ErrMsgKeyLen)
		}
		msg.Key = string(b[REPL_MSG_LEN_MIN:keyEnd])
	}

	// +END is already stripped by scanner split func
	msg.Value = b[keyEnd:]

	return msg, nil
}

// Serializes the given Msg
func EncodeReplMsg(msg *ReplMsg) ([]byte, error) {
	var buf bytes.Buffer

	// Expires [0:16]
	expBytes := make([]byte, 16)
	if msg.Expires > 0 {
		binary.BigEndian.PutUint64(expBytes, uint64(msg.Expires))
	}
	_, err := buf.Write(expBytes)
	if err != nil {
		return nil, err
	}

	// Modified [16:32]
	modBytes := make([]byte, 16)
	if msg.Modified > 0 {
		binary.BigEndian.PutUint64(modBytes, uint64(msg.Modified))
	}
	_, err = buf.Write(modBytes)
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
