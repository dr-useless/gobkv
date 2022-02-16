package main

import (
	"errors"
	"strings"

	"github.com/dr-useless/gobkv/common"
)

type Store struct {
	Cfg   *Config
	Parts map[string]*Part
}

func (s *Store) Ping(args *common.Args, res *common.StatusReply) error {
	res.Status = common.StatusOk
	return nil
}

func (s *Store) Get(args *common.Args, res *common.ValueReply) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		return errors.New("unauthorized")
	}
	part := s.getClosestPart(args.Key)
	part.Mux.RLock()
	defer part.Mux.RUnlock()
	res.Value = part.Data[args.Key]
	return nil
}

func (s *Store) Set(args *common.Args, res *common.StatusReply) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		return errors.New("unauthorized")
	}
	res.Status = common.StatusOk
	part := s.getClosestPart(args.Key)
	part.Mux.Lock()
	defer part.Mux.Unlock()
	part.Data[args.Key] = args.Value
	part.MustWrite = true
	return nil
}

func (s *Store) Del(args *common.Args, res *common.StatusReply) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		return errors.New("unauthorized")
	}
	res.Status = common.StatusOk
	part := s.getClosestPart(args.Key)
	part.Mux.Lock()
	defer part.Mux.Unlock()
	delete(part.Data, args.Key)
	part.MustWrite = true
	return nil
}

func (s *Store) List(args *common.Args, res *common.KeysReply) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		return errors.New("unauthorized")
	}
	res.Keys = make([]string, 0)
	for _, part := range s.Parts {
		for k := range part.Data {
			if args.Key == "" || strings.HasPrefix(k, args.Key) {
				res.Keys = append(res.Keys, k)
				if args.Limit != 0 && len(res.Keys) >= args.Limit {
					return nil
				}
			}
		}
	}
	return nil
}
