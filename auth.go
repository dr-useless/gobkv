package main

import (
	"bufio"
	"errors"
	"net"
	"strings"
)

func handleAuth(c net.Conn, authSecret string) error {
	if authSecret == "" {
		return nil
	}

	data, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		return err
	}

	if authSecret != strings.TrimSuffix(data, "\n") {
		return errors.New("unauthorized")
	}

	return nil
}
