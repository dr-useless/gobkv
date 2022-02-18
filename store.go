package main

import (
	"strings"
	"sync"
)

// Exported struct for net/rpc calls
// Holds configuration for convenience
type Store struct {
	AuthSecret string
	Parts      map[string]*Part
}

// Get Slot for specified key
// from appropriate partition
func (s *Store) Get(key string) *Slot {
	part := s.getClosestPart(key)
	part.Mux.RLock()
	defer part.Mux.RUnlock()
	return part.Data[key]
}

// Set specified Slot
// in appropriate partition
func (s *Store) Set(key string, slot *Slot) {
	part := s.getClosestPart(key)
	part.Mux.Lock()
	part.Data[key] = slot
	part.MustWrite = true
	part.Mux.Unlock()
}

// Remove Slot with specified key
// from appropriate partition
func (s *Store) Del(key string) {
	part := s.getClosestPart(key)
	part.Mux.Lock()
	delete(part.Data, key)
	part.MustWrite = true
	part.Mux.Unlock()
}

// Concurrently search all parts
// for the keys with the given prefix
// returns list of matching keys
func (s *Store) List(prefix string) []string {
	keys := make([]string, 0)
	mux := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for _, part := range s.Parts {
		wg.Add(1)
		go func(part *Part, keys *[]string, wg *sync.WaitGroup, mux *sync.Mutex) {
			defer wg.Done()
			var partKeys []string
			if prefix == "" {
				// no prefix given, will return all keys
				// so allocate enough space
				partKeys = make([]string, len(part.Data))
				i := 0
				for k := range part.Data {
					partKeys[i] = k
					i++
				}
			} else {
				partKeys = make([]string, 0)
				for k := range part.Data {
					if strings.HasPrefix(k, prefix) {
						partKeys = append(partKeys, k)
					}
				}
			}
			if len(partKeys) > 0 {
				mux.Lock()
				*keys = append(*keys, partKeys...)
				mux.Unlock()
			}
		}(part, &keys, wg, mux)
	}
	wg.Wait()
	return keys
}
