package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/dr-useless/gobkv/store"
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
	st.EnsureParts(&cfg.Parts)
	go st.ScanForExpiredKeys(&cfg.Parts, cfg.ExpiryScanPeriod)

	// if ReplServer is defined, start as master
	if cfg.ReplMasterConfig.Address != "" {
		log.Println("configured as master")
		replMaster := store.ReplMaster{}
		go replMaster.Init(&cfg.ReplMasterConfig)
		st.ReplMaster = &replMaster
	} else if cfg.ReplClientConfig.Address != "" {
		log.Println("configured as replica")
		replClient := store.ReplClient{
			Dir: cfg.Dir,
		}
		replClient.Store = &st
		replClient.Init(&cfg.ReplClientConfig)
	}

	watchdog := Watchdog{
		store: &st,
		cfg:   &cfg,
	}
	// blocks until parts are ready
	watchdog.readFromPartFiles()

	listener, err := store.GetListener(
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
		w.writeAllParts()
		os.Exit(0)
	}
}
