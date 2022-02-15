package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"hash/fnv"
	"log"
	"os"
	"path"
	"sync"
)

const listFileName = "shards.gob"
const hashLen = 4

type Shard struct {
	Id        []byte
	Mux       *sync.RWMutex
	Data      map[string][]byte
	MustWrite bool
}

func (shard *Shard) writeToFile(shardName string, cfg *Config) {
	if !shard.MustWrite {
		return
	}
	shard.Mux.RLock()
	defer shard.Mux.RUnlock()
	fullPath := path.Join(cfg.ShardDir, shardName+".gob")
	file, err := os.Create(fullPath)
	if err != nil {
		log.Fatalf("failed to create shard file: %s\r\n", err)
	}
	defer file.Close()
	gob.NewEncoder(file).Encode(&shard.Data)
	shard.MustWrite = false
}

// ensures that shard files exist
func (s *Store) ensureShards() {
	if !s.Cfg.Persist {
		return
	}
	s.Shards = make(map[string]*Shard)
	listPath := path.Join(s.Cfg.ShardDir, listFileName)
	listFile, err := os.Open(listPath)
	if err != nil {
		log.Println("no shard list found, will create...")
		// make new list
		for i := 0; i < s.Cfg.ShardCount; i++ {
			shardId := make([]byte, hashLen)
			rand.Read(shardId)
			shardName := getShardName(shardId)
			s.Shards[shardName] = &Shard{
				Id:   shardId,
				Mux:  new(sync.RWMutex),
				Data: make(map[string][]byte),
			}
		}
		newListFile, err := os.Create(listPath)
		if err != nil {
			log.Fatalf("failed to create shard list, check directory exists: %s", s.Cfg.ShardDir)
		}
		shardNameList := s.getShardNameList()
		gob.NewEncoder(newListFile).Encode(shardNameList)
	} else {
		// decode list
		nameList := make([]string, 0)
		err := gob.NewDecoder(listFile).Decode(&nameList)
		if err != nil {
			log.Fatalf("failed to decode shard list: %s", err)
		}
		for _, name := range nameList {
			id, _ := getShardId(name)
			s.Shards[name] = &Shard{
				Id:   id,
				Mux:  new(sync.RWMutex),
				Data: make(map[string][]byte),
			}
		}
		log.Printf("initialised %v shards from list\r\n", len(s.Shards))
	}
}

func (s *Store) getClosestShard(key string) *Shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	keyHash := h.Sum(nil)
	var clShard *Shard
	var clD []byte
	for _, shard := range s.Shards {
		d := xorBytes(shard.Id, keyHash)
		if clD == nil || bytes.Compare(d, clD) < 0 {
			clShard = shard
			clD = d
		}
	}
	return clShard
}

func (s *Store) getShardNameList() []string {
	list := make([]string, 0)
	for name := range s.Shards {
		list = append(list, name)
	}
	return list
}

func getShardName(shardId []byte) string {
	return base64.URLEncoding.EncodeToString(shardId)
}

func getShardId(shardName string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(shardName)
}
