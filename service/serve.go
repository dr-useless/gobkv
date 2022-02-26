package service

import (
	"bufio"
	"net"

	"github.com/intob/rocketkv/protocol"
	"github.com/intob/rocketkv/store"
)

func ServeConn(conn net.Conn, st *store.Store, authSecret string) {
	authed := authSecret == ""

	scan := bufio.NewScanner(conn)
	scan.Split(protocol.SplitPlusEnd)

loop:
	for scan.Scan() {
		mBytes := scan.Bytes()
		msg, err := protocol.DecodeMsg(mBytes)
		if err != nil {
			panic(err)
		}

		switch msg.Op {
		case protocol.OpPing:
			handlePing(conn)
			continue
		case protocol.OpAuth:
			authed = handleAuth(conn, msg, authSecret)
			continue
		default:
			if !authed {
				respond(conn, &protocol.Msg{
					Status: protocol.StatusUnauthorized,
				})
				break loop
			}
		}

		// requires auth
		switch msg.Op {
		case protocol.OpGet:
			handleGet(conn, msg, st)
		case protocol.OpSet:
			handleSet(conn, msg, st)
		case protocol.OpSetAck:
			handleSet(conn, msg, st)
		case protocol.OpDel:
			handleDel(conn, msg, st)
		case protocol.OpDelAck:
			handleDel(conn, msg, st)
		case protocol.OpList:
			handleList(conn, msg, st)
		case protocol.OpClose:
			break loop
		}
	}

	conn.Close()
}

func handlePing(conn net.Conn) {
	respond(conn, &protocol.Msg{
		Op:     protocol.OpPong,
		Status: protocol.StatusOk,
	})
}

func handleAuth(conn net.Conn, msg *protocol.Msg, secret string) bool {
	authed := msg.Key == secret
	if authed {
		respondWithStatus(conn, protocol.StatusOk)
	} else {
		respondWithStatus(conn, protocol.StatusUnauthorized)
	}
	return authed
}

func handleGet(conn net.Conn, msg *protocol.Msg, st *store.Store) {
	slot, found := st.Get(msg.Key)
	if !found {
		respondWithStatus(conn, protocol.StatusNotFound)
		return
	}
	respond(conn, &protocol.Msg{
		Status:  protocol.StatusOk,
		Key:     msg.Key,
		Value:   slot.Value,
		Expires: slot.Expires,
	})
}

func handleSet(conn net.Conn, msg *protocol.Msg, st *store.Store) {
	slot := store.Slot{
		Value:   msg.Value,
		Expires: msg.Expires,
	}
	st.Set(msg.Key, slot)
	if msg.Op == protocol.OpSetAck {
		respondWithStatus(conn, protocol.StatusOk)
	}
}

func handleDel(conn net.Conn, msg *protocol.Msg, st *store.Store) {
	st.Del(msg.Key)
	if msg.Op == protocol.OpDelAck {
		respondWithStatus(conn, protocol.StatusOk)
	}
}

// TODO: buffer keys
func handleList(conn net.Conn, msg *protocol.Msg, st *store.Store) {
	for k := range st.List(msg.Key, 100) {
		respond(conn, &protocol.Msg{
			Key: k,
		})
	}
	respondWithStatus(conn, protocol.StatusStreamEnd)
}

// TODO: handle errors
func respond(conn net.Conn, resp *protocol.Msg) {
	respEnc, err := protocol.EncodeMsg(resp)
	if err != nil {
		panic(err)
	}
	_, err = conn.Write(respEnc)
	if err != nil {
		panic(err)
	}
}

func respondWithStatus(conn net.Conn, status byte) {
	respond(conn, &protocol.Msg{
		Status: status,
	})
}
