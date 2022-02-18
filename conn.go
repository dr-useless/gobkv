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
func serveConn(conn net.Conn, store *Store, authSecret string) {
	backoff := BACKOFF // ms
	authed := authSecret == ""
	var msg protocol.Message
	var err error
loop:
	for {
		msg = protocol.Message{}
		err = msg.Read(conn)
		if err != nil {
			if backoff > BACKOFF_LIMIT {
				respondWithStatus(conn, protocol.StatusError)
				log.Println("conn timed out:", err)
				break
			}
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff *= 2
			continue
		}

		switch msg.Op {
		case protocol.OpPing:
			handlePing(conn)
			continue
		case protocol.OpAuth:
			authed = handleAuth(conn, &msg, authSecret)
			continue
		default:
			if !authed {
				respondWithStatus(conn, protocol.StatusUnauthorized)
				break loop
			}
		}

		// requires auth
		switch msg.Op {
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

func handleAuth(conn net.Conn, msg *protocol.Message, secret string) bool {
	authed := msg.Key == secret
	resp := protocol.Message{
		Op: protocol.OpAuth,
	}
	if authed {
		resp.Status = protocol.StatusOk
	} else {
		resp.Status = protocol.StatusUnauthorized
	}
	resp.Write(conn)
	return authed
}

func handleGet(conn net.Conn, msg *protocol.Message, store *Store) {
	slot := store.Get(msg.Key)
	resp := protocol.Message{
		Op:  protocol.OpGet,
		Key: msg.Key,
	}
	if slot != nil {
		resp.Status = protocol.StatusOk
		resp.Expires = slot.Expires
		resp.Value = slot.Value
	} else {
		resp.Status = protocol.StatusNotFound
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

func respondWithStatus(conn net.Conn, status byte) {
	resp := protocol.Message{
		Status: status,
	}
	resp.Write(conn)
}
