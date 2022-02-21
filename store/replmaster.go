package store

import (
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
	id         []byte
	inputChan  chan repl.ReplOp
	outputChan chan repl.ReplOp
}

type ReplMasterConfig struct {
	Size       int // max number of ops to buffer
	Network    string
	Address    string
	CertFile   string
	KeyFile    string
	AuthSecret string
}

func (r *ReplMaster) AddToHead(op repl.ReplOp) {
	// write to buffer of all registered clients
	for _, c := range r.clients {
		c.inputChan <- op
	}
}

func (r *ReplMaster) Init(cfg *ReplMasterConfig) {
	r.size = cfg.Size
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
			log.Println("repl client auth failed:", err)
			break
		}

		body, err := repl.DecodeReplClientMsg(msg.Body)
		if err != nil {
			log.Println("failed to decode repl client msg:", err)
			break
		}

		if body.AuthSecret != authSecret {
			break
		}

		if body.Head < r.tail {
			// full resync required
		} else {
			r.registerClient(&ReplClientReg{
				id: body.Id,
			})
		}
	}
	conn.Close()
}

func (r *ReplMaster) registerClient(reg *ReplClientReg) {
	key := base64.RawStdEncoding.EncodeToString(reg.id)
	reg.inputChan = make(chan repl.ReplOp)
	reg.outputChan = make(chan repl.ReplOp, r.size)
	r.mutex.Lock()
	r.clients[key] = reg
	r.mutex.Unlock()
}
