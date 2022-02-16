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
	store *Store
	cfg   *Config
}

// While watch() only takes care of writing to partitions,
// will only watch if persistence is enabled
func (w *Watchdog) watch() {
	if !w.cfg.Persist {
		return
	}
	period := w.cfg.PartWritePeriod
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
		part.writeToFile(name, w.cfg.PartDir)
	}
}

func (w *Watchdog) readFromPartFiles() {
	if !w.cfg.Persist {
		return
	}
	wg := new(sync.WaitGroup)
	for name, part := range w.store.Parts {
		wg.Add(1)
		go func(name string, part *Part, wg *sync.WaitGroup) {
			defer wg.Done()
			part.Mux.Lock()
			defer part.Mux.Unlock()
			fullPath := path.Join(w.cfg.PartDir, name+".gob")
			file, err := os.Open(fullPath)
			if err != nil {
				log.Printf("failed to open partition %s\r\n", name)
				return
			}
			err = gob.NewDecoder(file).Decode(&part.Data)
			if err != nil {
				log.Printf("failed to decode data in partition %s\r\n", name)
				return
			}
			log.Printf("read from partition %s", name)
		}(name, part, wg)
	}
	wg.Wait()
}
