package store

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

func GetConn(network, address, certFile, keyFile string) (net.Conn, error) {
	if certFile == "" {
		conn, err := net.Dial(network, address)
		if err != nil {
			log.Printf("failed to connect to %s over %s\r\n", address, network)
		}
		return conn, err
	} else {
		// load cert & key
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Printf("failed to load key pair: %s\r\n", err)
			return nil, err
		}
		config := tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		}
		// return client on tls connection
		conn, err := tls.Dial(network, address, &config)
		if err != nil {
			log.Printf("failed to connect to %s with TLS\r\n", address)
		}
		return conn, err
	}
}
