package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/intob/rocketkv/service"
	"github.com/intob/rocketkv/store"
)

var configFile = flag.String("c", "", "must be a file path")

func main() {
	flag.Parse()

	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("failed to load config")
		panic(err)
	}

	st := store.Store{
		Dir: cfg.Dir,
	}
	st.EnsureManifest(&cfg.Parts)
	go st.ScanForExpiredKeys(cfg.ExpiryScanPeriod)

	watchdog := Watchdog{
		store: &st,
		cfg:   &cfg,
	}
	// blocks until parts are ready
	watchdog.readFromBlockFiles()

	listener, err := service.GetListener(
		cfg.Network, cfg.Address, cfg.CertFile, cfg.KeyFile)
	if err != nil {
		panic(err)
	}

	go waitForSigInt(listener, &watchdog)
	go watchdog.watch()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go service.ServeConn(conn, &st, cfg.AuthSecret, cfg.BufferSize)
	}
}

func waitForSigInt(listener net.Listener, w *Watchdog) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		fmt.Println("will exit cleanly")
		listener.Close()
		w.writeAllBlocks()
		os.Exit(0)
	}
}
