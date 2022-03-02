package repl

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/intob/rocketkv/protocol"
	"github.com/intob/rocketkv/store"
	"github.com/intob/rocketkv/util"
)

type ReplService struct {
	cfg   ReplConfig
	store *store.Store
	peers map[uint64]*PeerConn
}

type PeerConn struct {
	Conn net.Conn
	Id   []byte
}

// TODO: figure out a reconnect/recover mechanism
func NewReplService(cfg ReplConfig, store *store.Store) *ReplService {
	svc := &ReplService{
		cfg:   cfg,
		store: store,
		peers: make(map[uint64]*PeerConn),
	}
	svc.cfg.Id = util.HashKey(cfg.Name)

	fmt.Println("my repl id:", util.GetNumber(svc.cfg.Id))

	go svc.startListener()

	// connect to peers with higher id than mine
	// peers with a lower ID will connect to me
	myIdNumber := util.GetNumber(svc.cfg.Id)
	for _, peer := range cfg.Peers {
		peer.Id = util.HashKey(peer.Name)
		peerIdNumber := util.GetNumber(peer.Id)
		if peerIdNumber > myIdNumber {
			fmt.Println("connecting to ", peer.Address)
			conn, err := util.GetConn(peer.Network, peer.Address, cfg.CertFile, cfg.KeyFile)
			if err != nil {
				fmt.Println("failed to connect to repl peer:", err)
				continue
			}

			pc := &PeerConn{
				Conn: conn,
				Id:   peer.Id,
			}

			ready := make(chan bool)
			go svc.writeConn(conn, pc, ready)
			<-ready

			go svc.readConn(conn, false)
		}

	}

	return svc
}

func (rs *ReplService) startListener() {
	c := &rs.cfg
	listener, err := util.GetListener(c.Network, c.Address, c.CertFile, c.KeyFile)
	if err != nil {
		fmt.Println("failed to start repl service listener:", err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go rs.readConn(conn, true)
	}
}

func (rs *ReplService) readConn(conn net.Conn, startWriter bool) {
	buf := make([]byte, rs.cfg.BufferSize)
	scan := bufio.NewScanner(conn)
	scan.Buffer(buf, cap(buf))
	scan.Split(protocol.SplitPlusEnd)

	authed := rs.cfg.AuthSecret == ""
	var peerReplId []byte
	var peerReplIdNum uint64

	if !scan.Scan() {
		conn.Close()
		return
	}

	// auth
	if !authed {
		auth := scan.Text()
		if auth != rs.cfg.AuthSecret {
			conn.Close()
			return
		}
	}

	peerReplId = scan.Bytes()
	if len(peerReplId) != util.ID_LEN {
		fmt.Println("got invalid peer repl id from", conn.RemoteAddr().String())
		conn.Close()
		return
	}

	peerReplIdNum = util.GetNumber(peerReplId)

	// not yet registered, add to peer conn map
	if rs.peers[peerReplIdNum] == nil {
		rs.peers[peerReplIdNum] = &PeerConn{
			Conn: conn,
			Id:   peerReplId,
		}
		if startWriter {
			ready := make(chan bool)
			go rs.writeConn(conn, rs.peers[peerReplIdNum], ready)
			<-ready
		}
	}

	for scan.Scan() {
		mBytes := scan.Bytes()
		// TODO: fix this
		if len(mBytes) < 1 {
			fmt.Println("empty message")
			continue
		}
		msg, err := DecodeReplMsg(mBytes)
		if err != nil {
			fmt.Println("failed to decode repl msg:", mBytes)
			panic(err)
		}

		if msg.Key == "" {
			fmt.Println("blank key")
			continue
		}

		slot := store.Slot{
			Value:    msg.Value,
			Expires:  msg.Expires,
			Modified: msg.Modified,
		}
		rs.store.Set(msg.Key, slot, true)
		fmt.Println("replicated ", msg.Key, " from ", peerReplIdNum)
	}

	conn.Close()
}

// Periodically write changed blocks
func (rs *ReplService) writeConn(conn net.Conn, peerConn *PeerConn, ready chan bool) {
	peerIdNumber := util.GetNumber(peerConn.Id)

	fmt.Println("writeConn:", peerIdNumber)

	buf := bufio.NewWriter(conn)

	// auth
	if rs.cfg.AuthSecret != "" {
		authMsg := rs.cfg.AuthSecret + protocol.SPLIT_MARKER
		_, err := buf.Write([]byte(authMsg))
		if err != nil {
			panic(err)
		}
	}

	// repl ID
	_, err := buf.Write(rs.cfg.Id)
	if err != nil {
		panic(err)
	}
	_, err = buf.Write([]byte(protocol.SPLIT_MARKER))
	if err != nil {
		panic(err)
	}

	buf.Flush()

	ready <- true

	//loop:
	for {
		for _, part := range rs.store.Parts {
			for _, block := range part.Blocks {
				block.Mutex.RLock()
				replState := block.ReplState[peerIdNumber]
				if replState == nil || replState.MustSync {
					for key, slot := range block.Slots {
						msg := &ReplMsg{
							Expires:  slot.Expires,
							Modified: slot.Modified,
							Key:      key,
							Value:    slot.Value,
						}
						mBytes, err := EncodeReplMsg(msg)
						if err != nil {
							fmt.Println("failed to encode repl block:", err)
							continue
						}
						buf.Write(mBytes)
						buf.Write([]byte(protocol.SPLIT_MARKER))
					}
					if replState == nil {
						block.ReplState[peerIdNumber] = &store.ReplNodeState{
							MustSync: false,
						}
					} else {
						block.ReplState[peerIdNumber].MustSync = false
					}
				}
				block.Mutex.RUnlock()
			}
		}
		buf.Flush()
		time.Sleep(time.Duration(rs.cfg.Period) * time.Second)
	}

	//conn.Close()
}
