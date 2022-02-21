package store

import (
	"strings"
	"sync"
	"time"
)

// Exported struct for net/rpc calls
// Holds configuration for convenience
type Store struct {
	Parts      map[string]*Part
	ReplServer *ReplServer
	Dir        string
}

// Get Slot for specified key
// from appropriate partition
func (s *Store) Get(key string) *Slot {
	part := s.getClosestPart(key)
	part.Mutex.RLock()
	defer part.Mutex.RUnlock()
	return part.Data[key]
}

// Set specified Slot
// in appropriate partition
func (s *Store) Set(key string, slot *Slot) {
	slot.Modified = uint64(time.Now().Unix())
	part := s.getClosestPart(key)
	part.Mutex.Lock()
	part.Data[key] = slot
	part.MustWrite = true
	part.Mutex.Unlock()
}

// Remove Slot with specified key
// from appropriate partition
func (s *Store) Del(key string) {
	part := s.getClosestPart(key)
	part.Mutex.Lock()
	delete(part.Data, key)
	part.MustWrite = true
	part.Mutex.Unlock()
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
				mutex.Lock()
				*keys = append(*keys, partKeys...)
				mutex.Unlock()
			}
		}(part, &keys, wg, mutex)
	}
	wg.Wait()
	return keys
}
