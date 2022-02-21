package store

// To do:
// Look at faster implementations
// https://go.dev/src/crypto/cipher/xor_generic.go
// Returns number of bytes xor'd
func xorBytes(a, b []byte) []byte {
	n := len(a)
	res := make([]byte, n)
	for i := 0; i < n; i++ {
		res[i] = a[i] ^ b[i]
	}
	return res
}
