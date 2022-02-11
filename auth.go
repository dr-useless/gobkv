package main

import (
	"bufio"
	"errors"
	"net"
	"net/textproto"
)

func handleAuth(conn net.Conn, authSecret string) error {
	if authSecret == "" {
		return nil
	}

	r := bufio.NewReader(conn)
	data, err := textproto.NewReader(r).ReadLine()
	if err != nil {
		return err
	}

	if authSecret != data {
		return errors.New("unauthorized")
	}

	return nil
}
