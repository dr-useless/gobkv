package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"log"
	"os"
	"os/signal"
	"path"
	"time"
)

type Watchdog struct {
	Store *Store
	Cfg   *Config
}

// While watch() only takes care of writing to shards,
// only watch if persistence is enabled
func (w *Watchdog) watch() {
	if !w.Cfg.Persist {
		return
	}
	log.Println("persistence enabled, will periodically write to fs")
	go w.waitForSigInt()
	for {
		w.writeIfMust()
		time.Sleep(time.Duration(10) * time.Second)
	}
}

func (w *Watchdog) waitForSigInt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		log.Println("will exit cleanly")
		w.writeIfMust()
		os.Exit(0)
	}
}

func (w *Watchdog) writeIfMust() {
	w.Store.Mux.Lock()
	defer w.Store.Mux.Unlock()
	for shardName, mustWrite := range w.Store.MustWrite {
		if mustWrite {
			w.writeShard(shardName)
		}
	}
}

// for now, collect shard data by ranging through all keys
// to do: shard data in memory also
// this will also help reduce blocking
// by having a separate mutex for each shard
func (w *Watchdog) writeShard(shardName string) {
	shard, _ := base64.URLEncoding.DecodeString(shardName)
	fullPath := path.Join(w.Cfg.ShardDir, shardName+".gob")
	file, err := os.Create(fullPath)
	if err != nil {
		log.Printf("failed to create shard file: %s\r\n", err)
	}
	defer file.Close()
	shardData := make(map[string][]byte)
	for key, value := range w.Store.Data {
		closest := w.Store.getClosestShard(key)
		if bytes.Equal(closest, shard) {
			shardData[key] = value
		}
	}
	gob.NewEncoder(file).Encode(&shardData)
	w.Store.MustWrite[shardName] = false
}

func (w *Watchdog) readFromShards() {
	if !w.Cfg.Persist {
		return
	}
	w.Store.Mux.Lock()
	defer w.Store.Mux.Unlock()
	for i, shard := range w.Store.Shards {
		name := getShardName(shard)
		fullPath := path.Join(w.Cfg.ShardDir, name+".gob")
		file, err := os.Open(fullPath)
		if err != nil {
			log.Printf("failed to open shard %v %s\r\n", i, name)
			continue
		}
		err = gob.NewDecoder(file).Decode(&w.Store.Data)
		if err != nil {
			log.Printf("failed to decode data in shard %v %s\r\n", i, name)
			continue
		}
		log.Printf("read from shard %v %s", i, name)
	}
}
