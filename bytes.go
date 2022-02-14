package main

import (
	"errors"
)

// To do:
// Look at faster implementations
// https://go.dev/src/crypto/cipher/xor_generic.go
// Returns number of bytes xor'd
func xorBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
	return n
}

// Returns true if a is less than b
func lessThan(a, b []byte) (bool, error) {
	if len(a) != len(b) {
		return false, errors.New("slices a & b must have the same length")
	}
	for i := 0; i < len(a); i++ {
		if a[i] < b[i] {
			return true, nil
		}
	}
	return false, nil
}
