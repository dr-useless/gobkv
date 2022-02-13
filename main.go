package main

import (
	"log"
	"net"
	"net/rpc"
	"sync"
)

func main() {
	log.SetPrefix("tcpkv ")

	// config
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("failed to load config ", err)
	}

	// store
	// TODO: load data from configured persistance
	store := Store{
		Data: make(map[string][]byte),
		Mux:  new(sync.RWMutex),
		Cfg:  &cfg,
	}
	rpc.Register(&store)

	l, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()

	log.Printf("listening on %s\r\n", cfg.Address)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
