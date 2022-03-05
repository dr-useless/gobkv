package util

import (
	"crypto/tls"
	"fmt"
	"net"
)

// GetConn gets a connection
func GetConn(network, address string) (net.Conn, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		fmt.Printf("failed to connect to %s over %s\r\n", address, network)
	}
	return conn, err
}

// GetConn gets a connection using TLS
func GetConnWithTLS(network, address, certFile, keyFile string) (net.Conn, error) {
	// load cert & key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		fmt.Printf("failed to load key pair: %s\r\n", err)
		return nil, err
	}
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	// return client on tls connection
	conn, err := tls.Dial(network, address, &config)
	if err != nil {
		fmt.Printf("failed to connect to %s with TLS\r\n", address)
	}
	return conn, err
}
