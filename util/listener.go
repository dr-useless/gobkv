package util

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
)

// GetListener gets a TCP listener
func GetListener(network, address string) (net.Listener, error) {
	fmt.Printf("listening on %s over %s\r\n", address, network)
	return net.Listen(network, address)
}

// GetListenerWithTLS loads the given cert & key & returns a listener
func GetListenerWithTLS(network, address, certFile, keyFile string) (net.Listener, error) {
	fmt.Printf(`
	listening on %s over %s\r\n
	expecting TLS connections\r\n`, address, network)

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
