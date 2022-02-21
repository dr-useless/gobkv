package store

import (
	"encoding/base64"
	"log"
	"math/rand"
	"time"
)

// When a Replica client connects,
// they will give their head offset.
// If their head is >= the master tail,
// they can consume the buffer to partially
// re-sync. If not, they must run a fully re-sync.
// A full resync can be done using List (we need to support streaming),
// then getting all keys.
// Inspired by Redis replication.
type ReplServer struct {
	id      []byte // randomly set on startup
	head    int    // offset of most recent item
	tail    int    // offset of oldest item
	clients []ReplClientReg
}

type ReplClientReg struct {
	id         []byte
	inputChan  chan ReplOp
	outputChan chan ReplOp
}

type ReplServerConfig struct {
	Size       int // max number of ops to buffer
	Network    string
	Address    string
	AuthSecret string
	CertFile   string
	KeyFile    string
}

func (r *ReplServer) AddToHead(op ReplOp) {
	// write to buffer of all registered clients
	for _, c := range r.clients {
		c.inputChan <- op
	}
}

func (r *ReplServer) Init(cfg *ReplServerConfig) {
	for _, c := range r.clients {
		c.inputChan = make(chan ReplOp)
		c.outputChan = make(chan ReplOp, cfg.Size)
	}
	r.id = make([]byte, 32)
	rand.Seed(time.Now().UnixMicro())
	rand.Read(r.id)
	log.Println("replication init as MASTER")
	log.Println("replication id:", base64.StdEncoding.EncodeToString(r.id))

	// set up listener for client connections

}
