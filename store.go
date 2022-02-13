package main

import (
	"sync"
)

const (
	StatusOk    = '_'
	StatusError = '!'
	StatusWarn  = '?'
)

type Store struct {
	Data map[string][]byte
	Mux  *sync.RWMutex
}

type Result struct {
	Status byte
	Value  []byte
	Keys   []string
}

type Args struct {
	Key   string
	Value []byte
}

func (s *Store) Get(args *Args, res *Result) error {
	s.Mux.RLock()
	defer s.Mux.RUnlock()
	res.Value = s.Data[args.Key]
	res.Status = StatusOk
	return nil
}

func (s *Store) Put(args *Args, res *Result) error {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	s.Data[args.Key] = args.Value
	res.Status = StatusOk
	return nil
}

func (s *Store) Del(args *Args, res *Result) error {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	delete(s.Data, args.Key)
	res.Status = StatusOk
	return nil
}
