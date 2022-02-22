package store

import (
	"strings"
	"sync"
	"time"

	"github.com/dr-useless/gobkv/protocol"
	"github.com/dr-useless/gobkv/repl"
)

// Exported struct for net/rpc calls
// Holds configuration for convenience
type Store struct {
	Parts      map[string]*Part
	ReplMaster *ReplMaster
	Dir        string
}

// Get Slot for specified key
// from appropriate partition
func (s *Store) Get(key string) *Slot {
	part := s.getClosestPart(key)
	part.Mutex.RLock()
	defer part.Mutex.RUnlock()
	return part.Slots[key]
}

// Set specified Slot
// in appropriate partition
func (s *Store) Set(key string, slot *Slot) {
	slot.Modified = time.Now().Unix()
	part := s.getClosestPart(key)
	part.Mutex.Lock()
	part.Slots[key] = slot
	part.MustWrite = true
	part.Mutex.Unlock()
	if s.ReplMaster != nil {
		s.ReplMaster.AddToHead(repl.Op{
			Op:       protocol.OpSet,
			Key:      key,
			Value:    slot.Value,
			Expires:  slot.Expires,
			Modified: slot.Modified,
		})
	}

}

// Remove Slot with specified key
// from appropriate partition
func (s *Store) Del(key string) {
	part := s.getClosestPart(key)
	part.Mutex.Lock()
	delete(part.Slots, key)
	part.MustWrite = true
	part.Mutex.Unlock()
	if s.ReplMaster != nil {
		s.ReplMaster.AddToHead(repl.Op{
			Op:  protocol.OpDel,
			Key: key,
		})
	}
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
				partKeys = make([]string, len(part.Slots))
				i := 0
				for k := range part.Slots {
					partKeys[i] = k
					i++
				}
			} else {
				partKeys = make([]string, 0)
				for k := range part.Slots {
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
