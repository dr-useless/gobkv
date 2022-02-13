package main

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

func getListener(cfg *Config) (net.Listener, error) {
	addr := fmt.Sprintf("0.0.0.0:%v", cfg.Port)
	log.Printf("listening on :%v\r\n", cfg.Port)
	if cfg.CertFile == "" {
		// no cert, return plain listener
		return net.Listen("tcp", addr)
	} else {
		log.Println("expecting TLS connections")
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			log.Fatalf("failed to load key pair: %s", err)
		}
		tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
		tlsConfig.Rand = rand.Reader
		return tls.Listen("tcp", addr, &tlsConfig)
	}
}
