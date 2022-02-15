package main

import (
	"encoding/gob"
	"log"
	"os"
	"os/signal"
	"path"
	"sync"
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
	period := w.Cfg.ShardWritePeriod
	if period == 0 {
		period = 10
	}
	log.Printf("will write changed shards every %v seconds\r\n", period)
	go w.waitForSigInt()
	for {
		w.writeAllShards()
		time.Sleep(time.Duration(period) * time.Second)
	}
}

func (w *Watchdog) waitForSigInt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		log.Println("will exit cleanly")
		w.writeAllShards()

		// pprof
		stopCPUProfile()
		makeMemProfile()

		os.Exit(0)
	}
}

func (w *Watchdog) writeAllShards() {
	for name, shard := range w.Store.Shards {
		shard.writeToFile(name, w.Cfg)
	}
}

func (w *Watchdog) readFromShardFiles() {
	if !w.Cfg.Persist {
		return
	}
	wg := new(sync.WaitGroup)
	for name, shard := range w.Store.Shards {
		wg.Add(1)
		go func(name string, shard *Shard, wg *sync.WaitGroup) {
			defer wg.Done()
			shard.Mux.Lock()
			defer shard.Mux.Unlock()
			fullPath := path.Join(w.Cfg.ShardDir, name+".gob")
			file, err := os.Open(fullPath)
			if err != nil {
				log.Printf("failed to open shard %s\r\n", name)
				return
			}
			err = gob.NewDecoder(file).Decode(&shard.Data)
			if err != nil {
				log.Printf("failed to decode data in shard %s\r\n", name)
				return
			}
			log.Printf("read from shard %s", name)
		}(name, shard, wg)
	}
	wg.Wait()
}
