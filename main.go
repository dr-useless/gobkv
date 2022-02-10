package main

import (
	"log"
	"net"
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
	// TODO: load from configured persistance
	store := Store{
		Data: make(map[string][]byte),
		Mux:  new(sync.RWMutex),
	}

	l, err := net.Listen("tcp4", cfg.Address)
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
			return
		}
		go handleConnection(conn, &cfg, &store)
	}
}
