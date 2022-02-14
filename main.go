package main

import (
	"log"
	"net/rpc"
)

func main() {
	log.SetPrefix("gobkv ")

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("failed to load config ", err)
	}

	store := Store{
		Cfg: &cfg,
	}
	store.ensureShards()

	watchdog := Watchdog{
		Store: &store,
		Cfg:   &cfg,
	}
	watchdog.readFromShardFiles()
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
