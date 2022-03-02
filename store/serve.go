package store

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/intob/rocketkv/protocol"
)

func ServeConn(conn net.Conn, st *Store, authSecret string, bufferSize int) {
	authed := authSecret == ""

	buf := make([]byte, bufferSize)
	scan := bufio.NewScanner(conn)
	scan.Buffer(buf, cap(buf))
	scan.Split(protocol.SplitPlusEnd)

loop:
	for scan.Scan() {
		mBytes := scan.Bytes()
		msg, err := protocol.DecodeMsg(mBytes)
		if err != nil {
			fmt.Println(err)
			break loop
		}

		switch msg.Op {
		case protocol.OpPing:
			err = handlePing(conn)
			if err != nil {
				break loop
			}
			continue
		case protocol.OpAuth:
			authed = handleAuth(conn, msg, authSecret)
			if !authed {
				break loop
			}
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
			err = handleGet(conn, msg, st)
		case protocol.OpSet:
			err = handleSet(conn, msg, st)
		case protocol.OpSetAck:
			err = handleSet(conn, msg, st)
		case protocol.OpDel:
			err = handleDel(conn, msg, st)
		case protocol.OpDelAck:
			err = handleDel(conn, msg, st)
		case protocol.OpList:
			err = handleList(conn, msg, st)
		case protocol.OpCount:
			err = handleCount(conn, msg, st)
		case protocol.OpClose:
			break loop
		}

		if err != nil {
			fmt.Println(err)
			break loop
		}
	}

	conn.Close()
}

func handlePing(conn net.Conn) error {
	return respond(conn, &protocol.Msg{
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

func handleGet(conn net.Conn, msg *protocol.Msg, st *Store) error {
	slot, found := st.Get(msg.Key)
	if !found {
		return respondWithStatus(conn, protocol.StatusNotFound)
	}
	return respond(conn, &protocol.Msg{
		Status:  protocol.StatusOk,
		Key:     msg.Key,
		Value:   slot.Value,
		Expires: slot.Expires,
	})
}

func handleSet(conn net.Conn, msg *protocol.Msg, st *Store) error {
	slot := Slot{
		Value:   msg.Value,
		Expires: msg.Expires,
	}
	st.Set(msg.Key, slot, false)
	if msg.Op == protocol.OpSetAck {
		return respondWithStatus(conn, protocol.StatusOk)
	}
	return nil
}

func handleDel(conn net.Conn, msg *protocol.Msg, st *Store) error {
	st.Del(msg.Key)
	if msg.Op == protocol.OpDelAck {
		return respondWithStatus(conn, protocol.StatusOk)
	}
	return nil
}

func handleList(conn net.Conn, msg *protocol.Msg, st *Store) error {
	buf := bufio.NewWriter(conn)
	for k := range st.List(msg.Key, 100) {
		enc, err := protocol.EncodeMsg(&protocol.Msg{
			Key: k,
		})
		if err != nil {
			return err
		}
		_, err = buf.Write(enc)
		if err != nil {
			return err
		}
	}
	err := buf.Flush()
	if err != nil {
		return err
	}
	return respondWithStatus(conn, protocol.StatusStreamEnd)
}

func handleCount(conn net.Conn, msg *protocol.Msg, st *Store) error {
	count := st.Count(msg.Key)
	countBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(countBytes, count)
	return respond(conn, &protocol.Msg{
		Op:     protocol.OpCount,
		Status: protocol.StatusOk,
		Key:    msg.Key,
		Value:  countBytes,
	})
}

func respond(conn net.Conn, resp *protocol.Msg) error {
	respEnc, err := protocol.EncodeMsg(resp)
	if err != nil {
		return err
	}
	_, err = conn.Write(respEnc)
	return err
}

func respondWithStatus(conn net.Conn, status byte) error {
	return respond(conn, &protocol.Msg{
		Status: status,
	})
}
