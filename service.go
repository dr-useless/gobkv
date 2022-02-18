package main

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
)

func getListener(cfg *Config) (net.Listener, error) {
	log.Printf("listening on %s\r\n", cfg.Address)
	if cfg.CertFile == "" {
		// no cert, return plain listener
		return net.Listen(cfg.Network, cfg.Address)
	} else {
		log.Println("expecting TLS connections")
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			log.Fatalf("failed to load key pair: %s", err)
		}
		tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
		tlsConfig.Rand = rand.Reader
		return tls.Listen(cfg.Network, cfg.Address, &tlsConfig)
	}
}
