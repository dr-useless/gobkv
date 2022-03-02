package store

import (
	"bytes"
	"sync"
	"time"

	"github.com/intob/rocketkv/util"
)

type Store struct {
	Parts  map[uint64]*Part
	Dir    string
	DelTtl time.Duration
}

// Get slot for specified key
// from appropriate partition
func (s *Store) Get(key string) (*Slot, bool) {
	h := util.HashKey(key)
	block := s.getClosestPart(h).getClosestBlock(h)
	block.Mutex.RLock()
	defer block.Mutex.RUnlock()
	slot, found := block.Slots[key]
	return &slot, found
}

// Set specified slot in appropriate block
func (s *Store) Set(key string, slot Slot, repl bool) {
	h := util.HashKey(key)
	block := s.getClosestPart(h).getClosestBlock(h)
	if repl && block.Slots[key].Modified > slot.Modified {
		// if key has been modified since, skip it
		return
	} else {
		slot.Modified = time.Now().Unix()
	}
	block.Mutex.Lock()
	block.Slots[key] = slot
	block.MustWrite = true
	// don't re-replicate (for now)
	// TODO: think more about this, maybe it's better
	// to re-replicate except to origin of repl.
	// This would ensure that all replicas arrive at a consistent state,
	// even if they are not all connected. However, it increases the ammount
	// of work that is done. Maybe we can make this a config option.
	if !repl {
		for _, replNodeState := range block.ReplState {
			if replNodeState != nil {
				replNodeState.MustSync = true
			}
		}
	}
	block.Mutex.Unlock()
}

// Remove slot with specified key
//
// So that other nodes can replicate this, without maintaining
// a list of deletes or a log of operations, simply set the
// expiry time. This allows replicas to follow before the key
// is cleaned up.
func (s *Store) Del(key string) {
	h := util.HashKey(key)
	block := s.getClosestPart(h).getClosestBlock(h)
	block.Mutex.Lock()
	slot := block.Slots[key]
	slot.Expires = time.Now().Add(s.DelTtl).Unix()
	slot.Modified = time.Now().Unix()
	block.Slots[key] = slot
	block.MustWrite = true
	for _, replNodeState := range block.ReplState {
		if replNodeState != nil {
			replNodeState.MustSync = true
		}
	}
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
		util.FastXor(dist, keyHash, part.Id)
		if clDist == nil || bytes.Compare(dist, clDist) < 0 {
			clPart = part
			clDist = dist
		}
		// reset dist slice
		dist = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	}

	return clPart
}
