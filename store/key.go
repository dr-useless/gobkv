package store

import "hash/fnv"

const KEY_HASH_LEN = 16 // 128 bits

func hashKey(key string) []byte {
	h := fnv.New128a()
	h.Write([]byte(key))
	return h.Sum(nil)
}
