package main

import (
	"bufio"
	"log"
	"net"
	"strings"
)

const (
	CmdAuth  = 'a'
	CmdGet   = 'g'
	CmdPut   = 'p'
	CmdDel   = 'd'
	CmdClose = 'c'
)

func handleConnection(conn net.Conn, cfg *Config, store *Store) {
	log.Printf("serving %s\n", conn.RemoteAddr().String())

	var isAuthed bool
	if cfg.AuthSecret == "" {
		isAuthed = true
	}

	r := bufio.NewReader(conn)
	for {
		cmd, err := r.ReadByte()
		if err != nil {
			log.Println(err)
			conn.Close()
			return
		}

		if cmd == CmdAuth {
			isAuthed = handleAuth(r, cfg.AuthSecret)
			if isAuthed {
				log.Println(conn.RemoteAddr().String(), "authed")
			} else {
				log.Println(conn.RemoteAddr().String(), "unauthorized")
			}
			continue
		}

		if isAuthed {
			switch cmd {
			case CmdGet:
				handleGet(r, conn, store)
			case CmdPut:
				handlePut(r, conn, store)
			case CmdDel:
				handleDel(r, conn, store)
			case CmdClose:
				conn.Close()
				return
			default:
				log.Println("unrecognized command")
				res := Result{
					Status: StatusError,
				}
				res.WriteOn(conn)
			}
			r.Reset(conn)
		} else {
			log.Println(conn.RemoteAddr().String(), "unauthorized")
		}
	}
}

func handleGet(r *bufio.Reader, conn net.Conn, store *Store) {
	key, err := r.ReadString('\n')
	if err != nil {
		res := Result{
			Status: StatusError,
		}
		res.WriteOn(conn)
		return
	}
	key = strings.TrimRight(key, "\t\r\n")
	res := Result{
		Status: StatusOk,
		Value:  store.get(key),
	}
	res.WriteOn(conn)
}

func handlePut(r *bufio.Reader, conn net.Conn, store *Store) {
	key, err := r.ReadString('\n')
	if err != nil {
		res := Result{
			Status: StatusError,
		}
		res.WriteOn(conn)
		return
	}
	key = strings.TrimRight(key, "\t\r\n")
	value, err := r.ReadBytes('\n')
	if err != nil {
		res := Result{
			Status: StatusError,
		}
		res.WriteOn(conn)
		return
	}
	value = value[:len(value)-1]
	store.put(key, value)
	res := Result{
		Status: StatusOk,
	}
	res.WriteOn(conn)
}

func handleDel(r *bufio.Reader, conn net.Conn, store *Store) {
	key, err := r.ReadString('\n')
	if err != nil {
		res := Result{
			Status: StatusError,
		}
		res.WriteOn(conn)
		return
	}
	key = strings.TrimRight(key, "\t\r\n")
	store.del(key)
	res := Result{
		Status: StatusOk,
	}
	res.WriteOn(conn)
}
