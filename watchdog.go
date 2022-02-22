package main

import (
	"log"
	"sync"
	"time"

	"github.com/dr-useless/gobkv/store"
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
		w.writeAllParts()
		time.Sleep(time.Duration(period) * time.Second)
	}
}

func (w *Watchdog) writeAllParts() {
	for name, part := range w.store.Parts {
		part.WriteToFile(name, w.cfg.Dir)
	}
}

func (w *Watchdog) readFromPartFiles() {
	if !w.cfg.Parts.Persist {
		return
	}
	wg := new(sync.WaitGroup)
	for _, part := range w.store.Parts {
		wg.Add(1)
		go part.ReadFromFile(wg, w.cfg.Dir)
	}
	wg.Wait()
}
