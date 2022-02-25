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
//
// Returns channel for list of matching keys
func (s *Store) List(prefix string, bufferSize int) <-chan string {
	output := make(chan string, bufferSize)
	wg := new(sync.WaitGroup)
	for _, part := range s.Parts {
		wg.Add(1)
		go s.listKeysInPart(prefix, part, output, wg)
	}
	// close output chan when done
	go func(wg *sync.WaitGroup, o chan string) {
		wg.Wait()
		close(output)
	}(wg, output)
	return output
}

func (s *Store) listKeysInPart(pf string, p *Part, o chan string, wg *sync.WaitGroup) {
	for _, block := range p.Blocks {
		block.Mutex.RLock()
		for k := range block.Slots {
			if strings.HasPrefix(k, pf) {
				o <- k
			}
		}
		block.Mutex.RUnlock()
	}
	wg.Done()
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
