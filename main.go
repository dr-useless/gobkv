package main

import (
	"log"
	"net/rpc"
	"sync"
)

func main() {
	log.SetPrefix("gobkv ")

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("failed to load config ", err)
	}

	store := Store{
		Data:      make(map[string][]byte),
		Mux:       new(sync.RWMutex),
		Cfg:       &cfg,
		MustWrite: make(map[string]bool, cfg.ShardCount),
	}
	store.ensureShards()

	watchdog := Watchdog{
		Store: &store,
		Cfg:   &cfg,
	}
	watchdog.readFromShards()
	go watchdog.watch()

	rpc.Register(&store)

	listener, err := getListener(&cfg)
	if err != nil {
		log.Println(err)
		return
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
