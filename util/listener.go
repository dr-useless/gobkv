package util

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
)

func GetListener(network, address, certFile, keyFile string) (net.Listener, error) {
	fmt.Printf("listening on %s over %s\r\n", address, network)
	if certFile == "" {
		// no cert, return plain listener
		return net.Listen(network, address)
	} else {
		fmt.Println("expecting TLS connections")
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			fmt.Println("failed to load key pair")
			panic(err)
		}
		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		tlsConfig.Rand = rand.Reader
		return tls.Listen(network, address, &tlsConfig)
	}
}
