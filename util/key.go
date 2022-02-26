package util

import "hash/fnv"

// Exactly 16 byte IDs have been chosen
// for 2 reasons:
// - it's enough for very large data-sets
// - we take advantage of fastxor's Block() (for perf)
const ID_LEN = 16

// Uses Go's built-in fast hash func FNV 128a
//
// Takes a string for convenience
func HashKey(key string) []byte {
	h := fnv.New128a() // as per ID_LEN
	h.Write([]byte(key))
	return h.Sum(nil)
}
