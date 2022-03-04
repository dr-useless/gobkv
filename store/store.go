package store

import (
	"bytes"
	"path"
	"sync"
	"time"

	"github.com/intob/rocketkv/util"
)

type Store struct {
	Parts map[uint64]*Part
	Dir   string
}

// Get slot for specified key
// from appropriate partition
func (s *Store) Get(key string) (*Slot, bool) {
	ns, name := path.Split(key)
	h := hashKey(ns, name)
	block := s.getClosestPart(h).getClosestBlock(h)
	block.Mutex.RLock()
	defer block.Mutex.RUnlock()
	slot, found := block.Slots[name]
	return &slot, found
}

// Set specified slot in appropriate block
func (s *Store) Set(key string, slot Slot, repl bool) {
	ns, name := path.Split(key)
	h := hashKey(ns, name)
	block := s.getClosestPart(h).getClosestBlock(h)
	if repl && block.Slots[name].Modified > slot.Modified {
		// if key has been modified since, skip it
		return
	} else {
		slot.Modified = time.Now().Unix()
	}
	block.Mutex.Lock()
	block.Slots[name] = slot
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
// TODO: add key to list of deletes to replicate
func (s *Store) Del(key string) {
	ns, name := path.Split(key)
	h := hashKey(ns, name)
	block := s.getClosestPart(h).getClosestBlock(h)
	block.Mutex.Lock()
	delete(block.Slots, name)
	block.MustWrite = true
	for _, replNodeState := range block.ReplState {
		if replNodeState != nil {
			replNodeState.MustSync = true
		}
	}
	block.Mutex.Unlock()
}

// Returns channel for list of matching keys
//
// If a namespace is given only that namespace will be searched
func (s *Store) List(key string, bufferSize int) <-chan string {
	output := make(chan string, bufferSize)
	// split into namespace & path if given a path separator
	ns, name := path.Split(key)

	if ns == "" {
		// namespace is empty, search all parts
		wg := new(sync.WaitGroup)
		for _, part := range s.Parts {
			wg.Add(1)
			go func(part *Part) {
				part.listKeys(ns, name, output)
				wg.Done()
			}(part)
		}

		// close output chan when done
		go func() {
			wg.Wait()
			close(output)
		}()
	} else {
		go func() {
			// namespace is given, search only namespace part
			h := hashKey(ns, name)
			part := s.getClosestPart(h)
			part.listKeys(ns, name, output)
			close(output)
		}()
	}

	return output
}

func (s *Store) Count(key string) uint64 {
	// split into namespace & path if given a path separator
	ns, name := path.Split(key)
	if ns == "" {
		// namespace is empty, search all parts
		var count uint64
		mu := new(sync.Mutex)
		wg := new(sync.WaitGroup)
		for _, part := range s.Parts {
			wg.Add(1)
			go func(part *Part) {
				c := part.countKeys(name)
				mu.Lock()
				count += c
				mu.Unlock()
				wg.Done()
			}(part)
		}
		wg.Wait()
		return count
	} else {
		// search only given namespace
		h := hashKey(ns, name)
		part := s.getClosestPart(h)
		return part.countKeys(name)
	}
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

// Returns hash of key
//
// If key contains a path separator, the key is split
// into namespace & name (like dir & filename). In this
// case, only the namespace is hashed.
func hashKey(namespace, name string) []byte {
	// hash namespace if not empty
	var h []byte
	if namespace == "" {
		h = util.HashStr(name)
	} else {
		h = util.HashStr(namespace)
	}
	return h
}
