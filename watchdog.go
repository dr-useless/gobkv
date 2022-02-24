package main

import (
	"log"
	"sync"
	"time"

	"github.com/intob/gobkv/store"
)

type Watchdog struct {
	store *store.Store
	cfg   *Config
}

// While watch() only takes care of writing to partitions,
// will only watch if persistence is enabled
func (w *Watchdog) watch() {
	if !w.cfg.Parts.Persist {
		return
	}
	period := w.cfg.Parts.WritePeriod
	if period == 0 {
		period = 10
	}
	log.Printf("will write changed partitions every %v seconds\r\n", period)
	for {
		w.writeAllBlocks()
		time.Sleep(time.Duration(period) * time.Second)
	}
}

func (w *Watchdog) writeAllBlocks() {
	for _, part := range w.store.Parts {
		for _, block := range part.Blocks {
			block.WriteToFile(w.cfg.Dir)
		}
	}
}

func (w *Watchdog) readFromBlockFiles() {
	if !w.cfg.Parts.Persist {
		return
	}
	for _, part := range w.store.Parts {
		wg := new(sync.WaitGroup)
		for _, block := range part.Blocks {
			wg.Add(1)
			go block.ReadFromFile(wg, w.cfg.Dir)
		}
		wg.Wait()
	}
}
