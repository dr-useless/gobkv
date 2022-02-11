package main

import (
	"bufio"
	"log"
	"net"
	"net/textproto"
)

func handleConnection(conn net.Conn, cfg *Config, store *Store) {
	log.Printf("serving %s\n", conn.RemoteAddr().String())

	// authenticate
	err := handleAuth(conn, cfg.AuthSecret)
	if err != nil {
		log.Println(conn.RemoteAddr().String(), "unauthorized")
		result := Result{Status: StatusError}
		result.Write(conn)
		conn.Close()
		return
	}

	log.Println(conn.RemoteAddr().String(), "authorized")

	for {
		r := bufio.NewReader(conn)
		data, err := textproto.NewReader(r).ReadDotBytes()
		if err != nil {
			log.Println("end ", err)
			return
		}

		cmd, err := parseCmd(data)
		if err != nil {
			log.Println("bad cmd ", err)
			continue
		}

		log.Printf("exec %s, k: %s, v: %v", string(cmd.Op), cmd.Key, cmd.Value)

		result := store.exec(cmd)
		result.Write(conn)
	}
}
