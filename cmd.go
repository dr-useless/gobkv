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
	// first byte is the operation
	op := cmd[0]
	if !isValidOp(op) {
		return Cmd{}, errors.New("invalid operation")
	}

	log.Println("op ", string(op))

	// second byte is the key length
	keyLen := int(cmd[1])

	log.Println("keyLen ", keyLen)

	// get key using key length
	key := cmd[1 : 1+keyLen]

	log.Println("key ", string(key))

	var value []byte
	// value is remainder of cmd
	if len(cmd) > 2+keyLen {
		valStart := 1 + keyLen
		value = cmd[valStart:]
		log.Println("value ", string(value))
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
