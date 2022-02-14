package main

import (
	"errors"
)

// To do:
// Look at faster implementations
// https://go.dev/src/crypto/cipher/xor_generic.go
// Returns number of bytes xor'd
func xorBytes(a, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, errors.New("slices a & b must be of equal length")
	}
	res := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = a[i] ^ b[i]
	}
	return res, nil
}
