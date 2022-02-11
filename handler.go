package main

import (
	"bufio"
	"log"
	"net"
	"strings"
)

func handleConnection(conn net.Conn, cfg *Config, store *Store) {
	log.Printf("serving %s\n", conn.RemoteAddr().String())

	// authenticate
	err := handleAuth(conn, cfg.AuthSecret)
	if err != nil {
		log.Println(conn.RemoteAddr().String(), "unauthorized")
		conn.Write([]byte("unauthorized\r\n"))
		conn.Close()
		return
	}

	log.Println(conn.RemoteAddr().String(), "authorized")

	for {
		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println("end ", err)
			return
		}

		trimmed := strings.TrimSuffix(data, "\n")

		cmd, err := parseCmd([]byte(trimmed))
		if err != nil {
			log.Println("bad cmd ", err)
			continue
		}

		log.Println("exec ", string(cmd.Op), " key ", cmd.Key, " value ", string(cmd.Value))
		store.exec(cmd, conn)
	}
}
