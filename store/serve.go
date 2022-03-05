package store

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/intob/rocketkv/protocol"
)

// ServeConn handles reading & writing messages
// from & to a connection
func (st *Store) ServeConn(conn net.Conn, authSecret string, bufferSize int) {
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

		if !authed && msg.Op == protocol.OpAuth {
			authed = handleAuth(conn, msg, authSecret)
			if !authed {
				break loop
			}
			continue
		}

		err = st.handle(conn, msg, authed)
		if err != nil {
			break loop
		}
	}

	conn.Close()
}

// handle is the main handler for messages
func (st *Store) handle(conn net.Conn, msg *protocol.Msg, authed bool) error {
	switch msg.Op {
	case protocol.OpPing:
		return handlePing(conn)
	default: // gatekeeping
		if !authed {
			respond(conn, &protocol.Msg{
				Status: protocol.StatusUnauthorized,
			})
			return errors.New("unauthorized")
		}
	}

	// requires auth
	switch msg.Op {
	case protocol.OpGet:
		return handleGet(conn, msg, st)
	case protocol.OpSet:
		return handleSet(conn, msg, st)
	case protocol.OpSetAck:
		return handleSet(conn, msg, st)
	case protocol.OpDel:
		return handleDel(conn, msg, st)
	case protocol.OpDelAck:
		return handleDel(conn, msg, st)
	case protocol.OpList:
		return handleList(conn, msg, st)
	case protocol.OpCount:
		return handleCount(conn, msg, st)
	case protocol.OpClose:
		return errors.New("closed by client")
	default:
		return errors.New("unrecognized operation")
	}
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
