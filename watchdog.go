package main

import (
	"log"
	"os"
	"os/signal"
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
	go w.waitForSigInt()
	for {
		w.writeAllParts()
		time.Sleep(time.Duration(period) * time.Second)
	}
}

func (w *Watchdog) waitForSigInt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		log.Println("will exit cleanly")
		w.writeAllParts()

		// pprof
		stopCPUProfile()
		makeMemProfile()

		os.Exit(0)
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
