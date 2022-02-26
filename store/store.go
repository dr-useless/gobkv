package store

import (
	"bytes"
	"sync"
	"time"

	"github.com/intob/rocketkv/util"
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
func (s *Store) Get(key string) (*Slot, bool) {
	h := util.HashKey(key)
	block := s.getClosestPart(h).getClosestBlock(h)
	block.Mutex.RLock()
	defer block.Mutex.RUnlock()
	slot, found := block.Slots[key]
	return &slot, found
}

// Set specified Slot
// in appropriate block
func (s *Store) Set(key string, slot Slot) {
	slot.Modified = time.Now().Unix()
	h := util.HashKey(key)
	block := s.getClosestPart(h).getClosestBlock(h)
	block.Mutex.Lock()
	block.Slots[key] = slot
	block.MustWrite = true
	block.Mutex.Unlock()
}

// Remove Slot with specified key
// from appropriate partition
func (s *Store) Del(key string) {
	h := util.HashKey(key)
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
		go func(part *Part) {
			part.listKeys(prefix, output)
			wg.Done()
		}(part)
	}
	// close output chan when done
	go func() {
		wg.Wait()
		close(output)
	}()
	return output
}

func (s *Store) Count(prefix string) uint64 {
	var count uint64
	mu := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for _, part := range s.Parts {
		wg.Add(1)
		go func(part *Part) {
			c := part.countKeys(prefix)
			mu.Lock()
			count += c
			mu.Unlock()
			wg.Done()
		}(part)
	}
	wg.Wait()
	return count
}

// Returns pointer to part with least Hamming distance
// from given key hash
func (s *Store) getClosestPart(keyHash []byte) *Part {
	var clDist []byte // winning distance
	var clPart *Part  // winning part
	dist := make([]byte, util.ID_LEN)

	// range through parts to find closest
	for _, part := range s.Parts {
		fastxor.Block(dist, keyHash, part.Id)
		if clDist == nil || bytes.Compare(dist, clDist) < 0 {
			clPart = part
			clDist = dist
		}
		// reset dist slice
		dist = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	}

	return clPart
}
