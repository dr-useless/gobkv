package main

import (
	"bufio"
	"log"
	"net"
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
		data, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			log.Println("end ", err)
			return
		}

		log.Println("cmd ", data)

		cmd, err := parseCmd(data)
		if err != nil {
			log.Println("bad cmd ", err)
			continue
		}

		log.Println("exec ", cmd.Op, cmd.Key)
		store.exec(cmd, conn)

		/*
			_, err = conn.Write(data)
			if err != nil {
				log.Println("write error: ", err)
				return
			}
		*/
	}
}
