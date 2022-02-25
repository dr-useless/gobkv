package store

import (
	"bytes"
	"strings"
	"sync"
	"time"

	"github.com/lukechampine/fastxor"
)

// Exported struct for net/rpc calls
// Holds configuration for convenience
type Store struct {
	Parts map[uint64]*Part
	Dir   string
}

// Get Slot for specified key
// from appropriate partition
func (s *Store) Get(key string) Slot {
	h := hashKey(key)
	block := s.getClosestPart(h).getClosestBlock(h)
	block.Mutex.RLock()
	defer block.Mutex.RUnlock()
	return block.Slots[key]
}

// Set specified Slot
// in appropriate block
func (s *Store) Set(key string, slot Slot) {
	slot.Modified = time.Now().Unix()
	h := hashKey(key)
	block := s.getClosestPart(h).getClosestBlock(h)
	block.Mutex.Lock()
	block.Slots[key] = slot
	block.MustWrite = true
	block.Mutex.Unlock()
}

// Remove Slot with specified key
// from appropriate partition
func (s *Store) Del(key string) {
	h := hashKey(key)
	block := s.getClosestPart(h).getClosestBlock(h)
	block.Mutex.Lock()
	delete(block.Slots, key)
	block.MustWrite = true
	block.Mutex.Unlock()
}

// Concurrently search all parts
// for the keys with the given prefix
// returns list of matching keys
func (s *Store) List(prefix string) []string {
	keys := make([]string, 0)
	mutex := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for _, part := range s.Parts {
		wg.Add(1)
		go func(part *Part, keys *[]string, wg *sync.WaitGroup, mutex *sync.Mutex) {
			defer wg.Done()
			partKeys := make([]string, 0)
			for _, block := range part.Blocks {
				for k := range block.Slots {
					if strings.HasPrefix(k, prefix) {
						partKeys = append(partKeys, k)
					}
				}
			}
			if len(partKeys) > 0 {
				mutex.Lock()
				*keys = append(*keys, partKeys...)
				mutex.Unlock()
			}
		}(part, &keys, wg, mutex)
	}
	wg.Wait()
	return keys
}

func (s *Store) getClosestPart(keyHash []byte) *Part {
	var clDist []byte // smallest distance value seen
	var clPart *Part  // part with smallest distance
	dist := make([]byte, KEY_HASH_LEN)

	// range through parts to find closest
	for _, part := range s.Parts {
		fastxor.Bytes(dist, keyHash, part.Id)
		if clDist == nil || bytes.Compare(dist, clDist) < 0 {
			clPart = part
			clDist = dist
		}
	}

	return clPart
}
