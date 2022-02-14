package main

import (
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
	go w.waitForSigInt()
	for {
		w.writeAllShards()
		time.Sleep(time.Duration(10) * time.Second)
	}
}

func (w *Watchdog) waitForSigInt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		log.Println("will exit cleanly")
		w.writeAllShards()
		os.Exit(0)
	}
}

func (w *Watchdog) writeAllShards() {
	for _, shard := range w.Store.Shards {
		shard.writeToFile(w.Cfg)
	}
}

func (w *Watchdog) readFromShardFiles() {
	if !w.Cfg.Persist {
		return
	}
	for name, shard := range w.Store.Shards {
		shard.Mux.Lock()
		defer shard.Mux.Unlock()
		fullPath := path.Join(w.Cfg.ShardDir, name+".gob")
		file, err := os.Open(fullPath)
		if err != nil {
			log.Printf("failed to open shard %s\r\n", name)
			continue
		}
		err = gob.NewDecoder(file).Decode(&shard.Data)
		if err != nil {
			log.Printf("failed to decode data in shard %s\r\n", name)
			continue
		}
		log.Printf("read from shard %s", name)
	}
}
