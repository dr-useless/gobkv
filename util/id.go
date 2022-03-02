package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
)

func RandomId() ([]byte, error) {
	id := make([]byte, ID_LEN)
	_, err := rand.Read(id)
	return id, err
}

// Returns base64url encoding of id
//
// Used for block filename
func GetName(id []byte) string {
	return base64.RawURLEncoding.EncodeToString(id)
}

// Returns big endian uint64 of id
func GetNumber(id []byte) uint64 {
	return binary.BigEndian.Uint64(id)
}
