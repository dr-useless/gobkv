package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"net"
	"time"

	"github.com/dr-useless/gobkv/protocol"
	"github.com/dr-useless/gobkv/store"
)

const BACKOFF = 10        // ms
const BACKOFF_LIMIT = 100 // ms

// Listens for requests
// & sends response
func serveConn(conn net.Conn, store *store.Store, authSecret string) {
	backoff := BACKOFF // ms
	authed := authSecret == ""
	var msg protocol.Message
	var err error
loop:
	for {
		msg = protocol.Message{}
		_, err = msg.ReadFrom(conn)
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
	resp.WriteTo(conn)
}

func handleAuth(conn net.Conn, msg *protocol.Message, secret string) bool {
	given, err := msg.Body.ReadString('\n')
	if err != nil {
		return false
	}
	authed := given == secret
	if authed {
		respondWithStatus(conn, protocol.StatusOk)
	} else {
		respondWithStatus(conn, protocol.StatusUnauthorized)
	}
	return authed
}

func handleGet(conn net.Conn, msg *protocol.Message, s *store.Store) {
	d, err := decodeMsgData(msg)
	if err != nil {
		respondWithStatus(conn, protocol.StatusError)
	}
	slot := s.Get(d.Key)
	if slot == nil {
		respondWithStatus(conn, protocol.StatusNotFound)
		return
	}
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(slot)
	resp := protocol.Message{
		Op:     protocol.OpGet,
		Status: protocol.StatusOk,
		Body:   &buf,
	}
	resp.WriteTo(conn)
}

func handleSet(conn net.Conn, msg *protocol.Message, s *store.Store) {
	d, err := decodeMsgData(msg)
	if err != nil {
		respondWithStatus(conn, protocol.StatusError)
	}
	slot := store.Slot{
		Value:   d.Value,
		Expires: d.Expires,
	}
	s.Set(d.Key, &slot)
	if msg.Op == protocol.OpSetAck {
		respondWithStatus(conn, protocol.StatusOk)
	}
}

func handleDel(conn net.Conn, msg *protocol.Message, s *store.Store) {
	d, err := decodeMsgData(msg)
	if err != nil {
		respondWithStatus(conn, protocol.StatusError)
	}
	s.Del(d.Key)
	if msg.Op == protocol.OpDelAck {
		respondWithStatus(conn, protocol.StatusOk)
	}
}

// TODO: add ability to stream unknown length,
// then stream keys as they are found (buffered)
func handleList(conn net.Conn, msg *protocol.Message, s *store.Store) {
	d, err := decodeMsgData(msg)
	if err != nil {
		respondWithStatus(conn, protocol.StatusError)
	}
	dResp := &protocol.Data{
		Keys: s.List(d.Key),
	}
	dRespEnc, err := encodeMsgData(dResp)
	if err != nil {
		respondWithStatus(conn, protocol.StatusError)
	}
	resp := protocol.Message{
		Op:     protocol.OpList,
		Status: protocol.StatusOk,
		Body:   dRespEnc,
	}
	resp.WriteTo(conn)
}

func respondWithStatus(conn net.Conn, status byte) {
	resp := protocol.Message{
		Status: status,
	}
	resp.WriteTo(conn)
}

func decodeMsgData(msg *protocol.Message) (*protocol.Data, error) {
	d := protocol.Data{}
	err := gob.NewDecoder(msg.Body).Decode(&d)
	return &d, err
}

func encodeMsgData(d *protocol.Data) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(d)
	return &buf, err
}
