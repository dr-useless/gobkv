package util

import (
	"hash/fnv"
	"unsafe"
)

const ID_LEN = 16

// Uses Go's built-in fast hash func FNV, 128a version
//
// Takes a string for convenience
func HashKey(key string) []byte {
	h := fnv.New128a() // as per ID_LEN
	h.Write([]byte(key))
	return h.Sum(nil)
}

// Xor 16 bytes the fastest way
//
// From https://github.com/lukechampine/fastxor
// Stores (a xor b) in dst, where a, b, and dst all have length 16.
func FastXor(dst, a, b []byte) {
	// profiling indicates that for 16-byte blocks, the cost of a function
	// call outweighs the SSE/AVX speedup
	dw := (*[2]uintptr)(unsafe.Pointer(&dst[0]))
	aw := (*[2]uintptr)(unsafe.Pointer(&a[0]))
	bw := (*[2]uintptr)(unsafe.Pointer(&b[0]))
	dw[0] = aw[0] ^ bw[0]
	dw[1] = aw[1] ^ bw[1]
}

// For benchmark comparison
func SlowXor(dst, a, b []byte) {
	n := len(a)
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
}
