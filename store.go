package main

import (
	"errors"
	"strings"

	"github.com/dr-useless/gobkv/common"
)

type Store struct {
	Cfg    *Config
	Shards map[string]*Shard
}

func (s *Store) Get(args *common.Args, res *common.ValueReply) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		return errors.New("unauthorized")
	}
	shard := s.getClosestShard(args.Key)
	shard.Mux.RLock()
	defer shard.Mux.RUnlock()
	res.Value = shard.Data[args.Key]
	return nil
}

func (s *Store) Set(args *common.Args, res *common.StatusReply) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		return errors.New("unauthorized")
	}
	res.Status = common.StatusOk
	shard := s.getClosestShard(args.Key)
	shard.Mux.Lock()
	defer shard.Mux.Unlock()
	shard.Data[args.Key] = args.Value
	shard.MustWrite = true
	return nil
}

func (s *Store) Del(args *common.Args, res *common.StatusReply) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		return errors.New("unauthorized")
	}
	res.Status = common.StatusOk
	shard := s.getClosestShard(args.Key)
	shard.Mux.Lock()
	defer shard.Mux.Unlock()
	delete(shard.Data, args.Key)
	shard.MustWrite = true
	return nil
}

func (s *Store) List(args *common.Args, res *common.KeysReply) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		return errors.New("unauthorized")
	}
	res.Keys = make([]string, 0)
	for _, shard := range s.Shards {
		for k := range shard.Data {
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
