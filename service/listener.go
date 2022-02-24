package service

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
)

func GetListener(network, address, certFile, keyFile string) (net.Listener, error) {
	log.Printf("listening on %s over %s\r\n", address, network)
	if certFile == "" {
		// no cert, return plain listener
		return net.Listen(network, address)
	} else {
		log.Println("expecting TLS connections")
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Fatalf("failed to load key pair: %s", err)
		}
		tlsConfig := tls.Config{Certificates: []tls.Certificate{cert}}
		tlsConfig.Rand = rand.Reader
		return tls.Listen(network, address, &tlsConfig)
	}
}
