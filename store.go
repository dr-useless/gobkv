package main

import (
	"errors"
	"strings"
	"sync"

	"github.com/dr-useless/gobkv/common"
)

// Exported struct for net/rpc calls
// Holds configuration for convenience
type Store struct {
	AuthSecret string
	Parts      map[string]*Part
}

// Returns OK status code: '_'
// Useful to cleanly detect a connection issue
// without use of syscalls on client
func (s *Store) Ping(args *common.Args, res *common.StatusReply) error {
	res.Status = common.StatusOk
	return nil
}

// Get value for specified key
// from appropriate partition
func (s *Store) Get(args *common.Args, res *common.ValueReply) error {
	if args.AuthSecret != s.AuthSecret {
		return errors.New("unauthorized")
	}
	part := s.getClosestPart(args.Key)
	part.Mux.RLock()
	res.Value = part.Data[args.Key].Value
	res.Expires = part.Data[args.Key].Expires
	part.Mux.RUnlock()
	return nil
}

// Set specified key
// with given value
// in appropriate partition
func (s *Store) Set(args *common.Args, res *common.StatusReply) error {
	if args.AuthSecret != s.AuthSecret {
		return errors.New("unauthorized")
	}
	res.Status = common.StatusOk
	part := s.getClosestPart(args.Key)
	part.Mux.Lock()
	part.Data[args.Key] = Key{
		Value:   args.Value,
		Expires: args.Expires,
	}
	part.MustWrite = true
	part.Mux.Unlock()
	return nil
}

// Remove specified key
// from appropriate partition
func (s *Store) Del(args *common.Args, res *common.StatusReply) error {
	if args.AuthSecret != s.AuthSecret {
		return errors.New("unauthorized")
	}
	res.Status = common.StatusOk
	part := s.getClosestPart(args.Key)
	part.Mux.Lock()
	delete(part.Data, args.Key)
	part.MustWrite = true
	part.Mux.Unlock()
	return nil
}

// Concurrently search all parts
// for the keys with the given prefix
// returns list of matching keys
func (s *Store) List(args *common.Args, res *common.KeysReply) error {
	if args.AuthSecret != s.AuthSecret {
		return errors.New("unauthorized")
	}
	res.Keys = make([]string, args.Limit)
	mux := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for _, part := range s.Parts {
		wg.Add(1)
		go func(part *Part, keys []string, wg *sync.WaitGroup, mux *sync.Mutex) {
			defer wg.Done()
			var partKeys []string
			if args.Key == "" {
				// no prefix given
				// will return all keys
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
					if strings.HasPrefix(k, args.Key) {
						partKeys = append(partKeys, k)
					}
				}
			}
			if len(partKeys) > 0 {
				mux.Lock()
				res.Keys = append(res.Keys, partKeys...)
				mux.Unlock()
			}
		}(part, res.Keys, wg, mux)
	}
	wg.Wait()
	return nil
}
