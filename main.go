package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/intob/gobkv/service"
	"github.com/intob/gobkv/store"
)

var configFile = flag.String("c", "", "must be a file path")

func main() {
	flag.Parse()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	st := store.Store{
		Dir: cfg.Dir,
	}
	st.EnsureBlocks(&cfg.Parts)
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
		log.Fatal(err)
	}

	go waitForSigInt(listener, &watchdog)
	go watchdog.watch()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go serveConn(conn, &st, cfg.AuthSecret)
	}
}

func waitForSigInt(listener net.Listener, w *Watchdog) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		log.Println("will exit cleanly")
		listener.Close()
		w.writeAllBlocks()
		os.Exit(0)
	}
}
