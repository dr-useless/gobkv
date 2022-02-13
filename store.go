package main

import (
	"errors"
	"sync"

	"github.com/dr-useless/gobkv/common"
)

const (
	StatusOk    = '_'
	StatusError = '!'
	StatusWarn  = '?'
)

type Store struct {
	Data map[string][]byte
	Mux  *sync.RWMutex
	Cfg  *Config
}

func (s *Store) Get(args *common.Args, res *common.Result) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		res.Status = StatusError
		return errors.New("unauthorized")
	}
	s.Mux.RLock()
	defer s.Mux.RUnlock()
	res.Value = s.Data[args.Key]
	res.Status = StatusOk
	return nil
}

func (s *Store) Put(args *common.Args, res *common.Result) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		res.Status = StatusError
		return errors.New("unauthorized")
	}
	s.Mux.Lock()
	defer s.Mux.Unlock()
	s.Data[args.Key] = args.Value
	res.Status = StatusOk
	return nil
}

func (s *Store) Del(args *common.Args, res *common.Result) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		res.Status = StatusError
		return errors.New("unauthorized")
	}
	s.Mux.Lock()
	defer s.Mux.Unlock()
	delete(s.Data, args.Key)
	res.Status = StatusOk
	return nil
}
