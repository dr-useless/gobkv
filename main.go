package main

import (
	"log"
	"net"
)

func main() {
	log.SetPrefix("tcpkv ")

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("failed to load config ", err)
	}

	l, err := net.Listen("tcp4", cfg.Address)
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()

	log.Printf("listening on %s\r\n", cfg.Address)

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go handleConnection(c)
	}
}
