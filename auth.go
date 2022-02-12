package main

import (
	"bufio"
	"log"
	"strings"
)

func handleAuth(r *bufio.Reader, authSecret string) bool {
	if authSecret == "" {
		return true
	}

	data, err := r.ReadString('\n')
	if err != nil {
		log.Println(err)
		return false
	}
	data = strings.TrimRight(data, "\t\r\n")

	return authSecret == data
}
