package main

import (
	"log"
	"net"

	"github.com/dr-useless/gobkv/protocol"
)

// Listens for requests
// & sends responses
func serveConn(conn net.Conn, store *Store) {
	//for {
	req := protocol.Message{}
	req.Read(conn)

	log.Printf("OP: %s\r\n", protocol.MapOp()[req.Op])

	switch req.Op {
	case protocol.OpPing:
		handlePing(conn)
	default:
		log.Println("unrecognized op")
		respondWithError(conn)
	}
	//}
}

func handlePing(conn net.Conn) {
	msg := protocol.Message{
		Op:     protocol.OpPong,
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
