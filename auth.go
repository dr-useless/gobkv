package main

import (
	"bufio"
	"log"
	"net"
	"strings"
)

func handleAuth(conn net.Conn, authSecret string) bool {
	if authSecret == "" {
		return true
	}

	r := bufio.NewReader(conn)
	data, err := r.ReadString('\n')
	if err != nil {
		log.Println(err)
		return false
	}
	data = strings.TrimRight(data, "\t\r\n")

	return authSecret == data
}
