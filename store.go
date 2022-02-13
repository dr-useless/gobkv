package main

import (
	"errors"
	"strings"
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

func (s *Store) List(args *common.Args, res *common.Result) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		res.Status = StatusError
		return errors.New("unauthorized")
	}
	if args.Limit != 0 {
		limit := min(args.Limit, len(s.Data))
		res.Keys = make([]string, 0, limit)
	} else {
		res.Keys = make([]string, 0)
	}
	for k := range s.Data {
		if strings.HasPrefix(k, args.Key) {
			res.Keys = append(res.Keys, k)
			if args.Limit != 0 && len(res.Keys) >= args.Limit {
				return nil
			}
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
