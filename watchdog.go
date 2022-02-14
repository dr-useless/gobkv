package main

import (
	"encoding/gob"
	"log"
	"os"
	"os/signal"
	"time"
)

type Watchdog struct {
	Store *Store
	Cfg   *Config
}

func (w *Watchdog) watch() {
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
	w.Store.Mux.RLock()
	mustWrite := w.Store.MustWrite
	w.Store.Mux.RUnlock()
	if mustWrite {
		w.writeToFile()
	}
}

func (w *Watchdog) writeToFile() {
	if w.Cfg.PersistFile == "" {
		return
	}
	file, err := os.Create(w.Cfg.PersistFile)
	if err != nil {
		log.Printf("failed to create persistence file: %s\r\n", err)
	}
	defer file.Close()
	w.Store.Mux.Lock()
	defer w.Store.Mux.Unlock()
	gob.NewEncoder(file).Encode(&w.Store.Data)
	w.Store.MustWrite = false
}

func (w *Watchdog) readFromFile() {
	if w.Cfg.PersistFile == "" {
		return
	}
	file, err := os.Open(w.Cfg.PersistFile)
	if err != nil {
		log.Printf("failed to open persistence file: %s\r\n", err)
		return
	}
	w.Store.Mux.Lock()
	defer w.Store.Mux.Unlock()
	err = gob.NewDecoder(file).Decode(&w.Store.Data)
	if err != nil {
		log.Printf("failed to decode persistence file: %s\r\n", err)
	}
}
