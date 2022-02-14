package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"log"
	"os"
	"path"

	"github.com/twmb/murmur3"
)

const listFileName = "shards.gob"
const hashLen = 32

// ensures that shard files exist
func (s *Store) ensureShards() {
	if !s.Cfg.Persist {
		return
	}
	fullListFileName := path.Join(s.Cfg.ShardDir, listFileName)
	listFile, err := os.Open(fullListFileName)
	if err != nil {
		log.Println("no shard list found, will create...")
		// make new list
		for i := 0; i < s.Cfg.ShardCount; i++ {
			randBytes := make([]byte, hashLen)
			rand.Read(randBytes)
			s.Shards = append(s.Shards, randBytes)
		}
		newListFile, err := os.Create(fullListFileName)
		if err != nil {
			log.Fatalf("failed to create shard list: %s", err)
		}
		gob.NewEncoder(newListFile).Encode(&s.Shards)
	} else {
		// decode list
		err := gob.NewDecoder(listFile).Decode(&s.Shards)
		if err != nil {
			log.Fatalf("failed to decode shard list: %s", err)
		}
		log.Println("loaded shard list")
		if len(s.Shards) != s.Cfg.ShardCount {
			log.Fatal("configured shard count does not match shard list")
		}
	}
}

func (s *Store) getClosestShard(key string) []byte {
	h := murmur3.New32()
	h.Write([]byte(key))
	keyHash := h.Sum(nil)

	// start with first shard
	clShard := s.Shards[0]
	clD := make([]byte, hashLen)
	xorBytes(clD, clShard, keyHash)

	// compare first with the rest
	for _, shard := range s.Shards[1:] {
		d := make([]byte, hashLen)
		xorBytes(d, shard, keyHash)
		isCloser, _ := lessThan(d, clD)
		if isCloser {
			clShard = shard
			clD = d
		}
	}

	return clShard
}

func getShardName(shard []byte) string {
	return base64.URLEncoding.EncodeToString(shard)
}
