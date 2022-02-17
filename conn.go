package main

import (
	"log"
	"net"

	"github.com/dr-useless/gobkv/protocol"
)

// Listens for requests
// & sends responses
func serveConn(conn net.Conn) {
	for {
		req := protocol.Message{}
		req.Read(conn)

		log.Println("msg", req)

		switch req.Op {
		case protocol.OpPing:
			handlePing(conn)
		default:
			log.Println("unrecognized op")
			respondWithError(conn)
		}
	}
}

func handlePing(conn net.Conn) {
	msg := protocol.Message{
		Status: protocol.StatusOk,
	}
	msg.Write(conn)
}

func respondWithError(conn net.Conn) {
	msg := protocol.Message{
		Status: protocol.StatusError,
	}
	msg.Write(conn)
}
