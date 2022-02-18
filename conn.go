package main

import (
	"log"
	"net"
	"strings"
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
	var msg protocol.Message
	var err error
loop:
	for {
		msg = protocol.Message{}
		err = msg.Read(conn)
		if err != nil {
			log.Println(err)
			if backoff > BACKOFF_LIMIT {
				respondWithError(conn)
				log.Println("conn timed out", err)
				break
			}
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff *= 2
			continue
		}
		ops++

		switch msg.Op {
		case protocol.OpPing:
			handlePing(conn)
		case protocol.OpGet:
			handleGet(conn, &msg, store)
		case protocol.OpSet:
			handleSet(conn, &msg, store)
		case protocol.OpSetAck:
			handleSet(conn, &msg, store)
		case protocol.OpDel:
			handleDel(conn, &msg, store)
		case protocol.OpDelAck:
			handleDel(conn, &msg, store)
		case protocol.OpList:
			handleList(conn, &msg, store)
		case protocol.OpClose:
			break loop
		default:
			log.Println("unrecognized op", msg.Op)
			respondWithError(conn)
			break loop
		}
	}

	log.Println("ops: ", ops)
	conn.Close()
}

func handlePing(conn net.Conn) {
	resp := protocol.Message{
		Op:     protocol.OpPong,
		Status: protocol.StatusOk,
	}
	resp.Write(conn)
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

func handleSet(conn net.Conn, msg *protocol.Message, store *Store) {
	slot := Slot{
		Value:   msg.Value,
		Expires: msg.Expires,
	}
	store.Set(msg.Key, &slot)
	if msg.Op == protocol.OpSetAck {
		resp := protocol.Message{
			Op:     msg.Op,
			Status: protocol.StatusOk,
		}
		err := resp.Write(conn)
		if err != nil {
			log.Println(err)
		}
	}
}

func handleDel(conn net.Conn, msg *protocol.Message, store *Store) {
	store.Del(msg.Key)
	if msg.Op == protocol.OpDelAck {
		resp := protocol.Message{
			Op:     msg.Op,
			Status: protocol.StatusOk,
			Key:    msg.Key,
		}
		resp.Write(conn)
	}
}

// TODO: add ability to stream unknown length,
// then stream keys as they are found (buffered)
func handleList(conn net.Conn, msg *protocol.Message, store *Store) {
	keys := store.List(msg.Key)
	keyStr := strings.Join(keys, " ")
	resp := protocol.Message{
		Op:     protocol.OpList,
		Status: protocol.StatusOk,
		Value:  []byte(keyStr),
	}
	resp.Write(conn)
}

func respondWithError(conn net.Conn) {
	resp := protocol.Message{
		Status: protocol.StatusError,
	}
	resp.Write(conn)
}
