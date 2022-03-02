package util

import (
	"crypto/tls"
	"fmt"
	"net"
)

// Get a connection with/without TLS
//
// TODO: remove InsecureSkipVerify from tls.Config
func GetConn(network, address, certFile, keyFile string) (net.Conn, error) {
	if certFile == "" {
		conn, err := net.Dial(network, address)
		if err != nil {
			fmt.Printf("failed to connect to %s over %s\r\n", address, network)
		}
		return conn, err
	} else {
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
}
