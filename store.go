package main

import (
	"encoding/gob"
	"errors"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/dr-useless/gobkv/common"
)

type Store struct {
	Data map[string][]byte
	Mux  *sync.RWMutex
	Cfg  *Config
}

func (s *Store) Get(args *common.Args, res *common.ValueReply) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		return errors.New("unauthorized")
	}
	s.Mux.RLock()
	defer s.Mux.RUnlock()
	res.Value = s.Data[args.Key]
	return nil
}

func (s *Store) Put(args *common.Args, res *common.StatusReply) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		return errors.New("unauthorized")
	}
	res.Status = common.StatusOk
	s.Mux.Lock()
	defer s.Mux.Unlock()
	s.Data[args.Key] = args.Value
	s.writeToFile()
	return nil
}

func (s *Store) Del(args *common.Args, res *common.StatusReply) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		return errors.New("unauthorized")
	}
	res.Status = common.StatusOk
	s.Mux.Lock()
	defer s.Mux.Unlock()
	delete(s.Data, args.Key)
	s.writeToFile()
	return nil
}

func (s *Store) List(args *common.Args, res *common.KeysReply) error {
	if args.AuthSecret != s.Cfg.AuthSecret {
		return errors.New("unauthorized")
	}
	if args.Limit != 0 {
		limit := min(args.Limit, len(s.Data))
		res.Keys = make([]string, 0, limit)
	} else {
		res.Keys = make([]string, 0)
	}
	for k := range s.Data {
		if args.Key == "" || strings.HasPrefix(k, args.Key) {
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

func (s *Store) writeToFile() {
	if s.Cfg.PersistFile == "" {
		return
	}
	file, err := os.Create(s.Cfg.PersistFile)
	if err != nil {
		log.Printf("failed to create persistence file: %s\r\n", err)
	}
	defer file.Close()
	gob.NewEncoder(file).Encode(&s.Data)
}

func (s *Store) readFromFile() {
	if s.Cfg.PersistFile == "" {
		return
	}
	file, err := os.Open(s.Cfg.PersistFile)
	if err != nil {
		log.Printf("failed to open persistence file: %s\r\n", err)
		return
	}
	err = gob.NewDecoder(file).Decode(&s.Data)
	if err != nil {
		log.Printf("failed to decode persistence file: %s\r\n", err)
	}
}
