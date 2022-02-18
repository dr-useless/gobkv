package main

import (
	"log"
	"net"
	"time"

	"github.com/dr-useless/gobkv/protocol"
)

const BACKOFF = 10        // ms
const BACKOFF_LIMIT = 100 // ms

// Listens for requests
// & sends response
func serveConn(conn net.Conn, store *Store) {
	backoff := BACKOFF // ms
	ops := 0
	for {
		msg := protocol.Message{}
		err := msg.Read(conn)
		if err != nil {
			log.Println(err)
			if backoff > BACKOFF_LIMIT {
				respondWithError(conn)
				log.Println("conn timed out, ops:", ops, err)
				break
			}
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff *= 2
			continue
		}
		backoff = BACKOFF
		ops++

		switch msg.Op {
		case protocol.OpPing:
			handlePing(conn)
		case protocol.OpGet:
			handleGet(conn, &msg, store)
		case protocol.OpSet:
			handleSet(conn, &msg, store)
		default:
			log.Println("unrecognized op", msg.Op)
			respondWithError(conn)
			conn.Close()
			return
		}
	}
	conn.Close()
}

func handlePing(conn net.Conn) {
	resp := protocol.Message{
		Op:     protocol.OpPong,
		Status: protocol.StatusOk,
	}
	resp.Write(conn)
}

func handleSet(conn net.Conn, msg *protocol.Message, store *Store) {
	slot := Slot{
		Value:   msg.Value,
		Expires: msg.Expires,
	}
	store.Set(msg.Key, &slot)
	resp := protocol.Message{
		Op:     protocol.OpGet,
		Status: protocol.StatusOk,
	}
	err := resp.Write(conn)
	if err != nil {
		log.Println(err)
	}
}

func handleGet(conn net.Conn, msg *protocol.Message, store *Store) {
	slot := store.Get(msg.Key)
	resp := protocol.Message{
		Op:      protocol.OpGet,
		Status:  protocol.StatusOk,
		Expires: slot.Expires,
		Key:     msg.Key,
		Value:   slot.Value,
	}
	resp.Write(conn)
}

func respondWithError(conn net.Conn) {
	resp := protocol.Message{
		Status: protocol.StatusError,
	}
	resp.Write(conn)
}
