package main

import (
	"errors"
	"log"
)

const (
	OpGet = 'g'
	OpPut = 'p'
	OpDel = 'd'
)

type Cmd struct {
	Op    byte
	Key   string
	Value []byte
}

// Takes a cmd byte slice,
// and returns a Cmd type
func parseCmd(cmd []byte) (Cmd, error) {
	if len(cmd) < 1 {
		return Cmd{}, errors.New("invalid operation")
	}

	// first byte is the operation
	op := cmd[0]
	if !isValidOp(op) {
		return Cmd{}, errors.New("invalid operation")
	}

	log.Println("op ", string(op))

	if len(cmd) < 2 {
		// op only
		return Cmd{
			Op: op,
		}, nil
	}

	// second byte is the key length
	keyLen := int(cmd[1])

	// validate key length
	if len(cmd) < 2+keyLen {
		return Cmd{}, errors.New("key length does not match")
	}

	// get key
	key := cmd[1 : 2+keyLen]

	// get value from remainder of cmd
	var value []byte
	valStart := 2 + keyLen
	if len(cmd) > valStart {
		value = cmd[valStart:]
	}

	return Cmd{
		Op:    op,
		Key:   string(key),
		Value: value,
	}, nil
}

func isValidOp(op byte) bool {
	switch op {
	case OpGet:
		return true
	case OpPut:
		return true
	case OpDel:
		return true
	default:
		return false
	}
}
