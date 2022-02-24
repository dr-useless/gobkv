package store

import (
	"bytes"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/intob/gobkv/protocol"
)

// Exported struct for net/rpc calls
// Holds configuration for convenience
type Store struct {
	Parts map[uint64]*Part
	Dir   string
}

// Get Slot for specified key
// from appropriate partition
func (s *Store) Get(key string) *Slot {
	block := s.getClosestPart(key).getClosestBlock(key)
	block.Mutex.RLock()
	defer block.Mutex.RUnlock()
	return block.Slots[key]
}

// Set specified Slot
// in appropriate block
func (s *Store) Set(key string, slot *Slot) {
	slot.Modified = time.Now().Unix()
	block := s.getClosestPart(key).getClosestBlock(key)
	block.Mutex.Lock()
	block.Slots[key] = slot
	block.MustWrite = true
	block.Mutex.Unlock()
}

// Remove Slot with specified key
// from appropriate partition
func (s *Store) Del(key string) {
	block := s.getClosestPart(key).getClosestBlock(key)
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

func (s *Store) getClosestPart(key string) *Part {
	h := fnv.New64a()
	h.Write([]byte(key))
	keyHash := h.Sum(nil)
	var clPart *Part
	var clD []byte
	for _, part := range s.Parts {
		d := xorBytes(part.Id, keyHash)
		if clD == nil || bytes.Compare(d, clD) < 0 {
			clPart = part
			clD = d
		}
	}
	return clPart
}

func (s *Store) getManifest() *protocol.Manifest {
	manifest := make(protocol.Manifest, 0)
	for _, part := range s.Parts {
		partManifest := protocol.PartManifest{
			PartId: part.Id,
			Blocks: make([]protocol.BlockManifest, 0),
		}
		for _, block := range part.Blocks {
			blockManifest := protocol.BlockManifest{
				BlockId: block.Id,
				Hash:    block.Checksum(),
			}
			partManifest.Blocks = append(partManifest.Blocks, blockManifest)
		}
		manifest = append(manifest, partManifest)
	}
	return &manifest
}
