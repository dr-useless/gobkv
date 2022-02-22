package store

import (
	"bytes"
	"encoding/base64"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/dr-useless/gobkv/protocol"
	"github.com/dr-useless/gobkv/repl"
)

// When a Replica client connects,
// they will give their head offset.
// If their head is >= the master tail,
// they can consume the buffer to partially
// re-sync. If not, they must run a fully re-sync.
// A full resync can be done using List (we need to support streaming),
// then getting all keys.
// Inspired by Redis replication.
type ReplMaster struct {
	id      []byte // randomly set on startup
	head    int    // offset of most recent item
	tail    int    // offset of oldest item
	size    int
	mutex   *sync.Mutex
	clients map[string]*ReplClientReg
}

type ReplClientReg struct {
	inputChan  chan repl.Op
	outputChan chan repl.Op
	conn       net.Conn
}

type ReplMasterConfig struct {
	Size       int // max number of ops to buffer
	Network    string
	Address    string
	CertFile   string
	KeyFile    string
	AuthSecret string
}

func (r *ReplMaster) AddToHead(op repl.Op) {
	// write to buffer of all registered clients
	for _, c := range r.clients {
		c.inputChan <- op
	}
	r.mutex.Lock()
	r.head++
	r.mutex.Unlock()
}

func (r *ReplMaster) Init(cfg *ReplMasterConfig) {
	r.size = cfg.Size
	r.clients = make(map[string]*ReplClientReg)
	r.mutex = new(sync.Mutex)
	r.id = make([]byte, 32)
	rand.Seed(time.Now().UnixMicro())
	rand.Read(r.id)
	log.Println("repl id:", base64.StdEncoding.EncodeToString(r.id))

	// set up listener for client connections
	listener, err := GetListener(cfg.Network, cfg.Address, cfg.CertFile, cfg.KeyFile)
	if err != nil {
		log.Fatal("failed to start repl listener:", err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go r.serveReplClient(conn, cfg.AuthSecret)
	}
}

func (r *ReplMaster) serveReplClient(conn net.Conn, authSecret string) {
	for {
		msg, err := protocol.ReadMsgFrom(conn)
		if err != nil {
			log.Println("repl client read err:", err)
			break
		}

		body := repl.ClientMsgBody{}
		err = body.DecodeFrom(msg.Body)
		if err != nil {
			log.Println("failed to decode repl client msg:", err)
			break
		}

		if body.AuthSecret != authSecret {
			break
		}

		respBody := repl.MasterMsgBody{
			ReplId: r.id,
			Head:   r.head,
		}
		if body.Head < r.tail || !bytes.Equal(body.ReplId, r.id) {
			// full sync required
			respBody.MustSync = true
		}

		respBodyEnc, _ := respBody.Encode()
		respMsg := protocol.Msg{
			Body: respBodyEnc,
		}
		respMsg.WriteTo(conn)

		r.registerClient(conn, body.ClientId)
	}
	conn.Close()
}

func (r *ReplMaster) registerClient(conn net.Conn, clientId []byte) {
	key := base64.RawStdEncoding.EncodeToString(clientId)
	r.mutex.Lock()
	reg := r.clients[key]
	if reg == nil {
		reg = &ReplClientReg{}
	}
	reg.conn = conn
	if reg.inputChan == nil {
		reg.inputChan = make(chan repl.Op)
		reg.outputChan = make(chan repl.Op, r.size)
	}
	r.clients[key] = reg
	r.mutex.Unlock()
	// keep buffer rolling
	go func(reg *ReplClientReg) {
		for op := range reg.inputChan {
			select {
			case reg.outputChan <- op:
			default:
				<-reg.outputChan
				reg.outputChan <- op
				r.mutex.Lock()
				r.tail++
				r.mutex.Unlock()
			}
		}
	}(reg)
	go r.pushOps(reg)
}

func (r *ReplMaster) pushOps(reg *ReplClientReg) {
	// send things
	go func(conn net.Conn, outputChan chan repl.Op) {
		errCount := 0
		for op := range outputChan {
			data, err := op.Encode()
			if err != nil {
				log.Println("failed to encode repl op:", err)
				continue
			}
			msg := protocol.Msg{
				Body: data,
			}
			_, err = msg.WriteTo(conn)
			if err != nil {
				log.Println("failed to send repl op:", err)
				errCount++
				if errCount > 5 {
					conn.Close()
					break
				}
				continue
			}
			errCount = 0
			log.Println("sent ReplOp to client, key:", op.Key)
		}
	}(reg.conn, reg.outputChan)
}
