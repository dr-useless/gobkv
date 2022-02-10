package main

import (
	"net"
	"sync"
)

type Store struct {
	Data map[string][]byte
	Mux  *sync.RWMutex
}

func (s *Store) get(key string) []byte {
	s.Mux.RLock()
	defer s.Mux.RUnlock()
	return s.Data[key]
}

func (s *Store) put(key string, value []byte) {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	s.Data[key] = value
}

func (s *Store) del(key string) {
	s.Mux.Lock()
	defer s.Mux.Unlock()
	delete(s.Data, key)
}

func (s *Store) exec(cmd Cmd, c net.Conn) {
	switch cmd.Op {
	case OpGet:
		value := s.get(cmd.Key)
		c.Write(value)
	case OpPut:
		s.put(cmd.Key, cmd.Value)
	case OpDel:
		s.del(cmd.Key)
	}
}