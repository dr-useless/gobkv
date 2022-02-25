package store

import "hash/fnv"

const KEY_HASH_LEN = 8 // 64 bits

func hashKey(key string) []byte {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum(nil)
}
