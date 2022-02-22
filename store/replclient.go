package store

import (
	"encoding/base64"
	"encoding/gob"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"time"

	"github.com/dr-useless/gobkv/protocol"
	"github.com/dr-useless/gobkv/repl"
)

const REPL_FILENAME = "repl.gob"
const BACKOFF = 10        // ms
const BACKOFF_LIMIT = 100 // ms

type ReplClient struct {
	State           *ReplClientState
	HeadIncremented bool
	Store           *Store
	Dir             string
}

type ReplClientState struct {
	Id   []byte
	Head int
}

type ReplClientConfig struct {
	Network    string
	Address    string
	CertFile   string
	KeyFile    string
	AuthSecret string
}

// spin up client & connect to master
func (rc *ReplClient) Init(cfg *ReplClientConfig) {
	rc.ensureStateFile()
	go rc.writeStateToFilePeriodically()

	conn, err := GetConn(cfg.Network, cfg.Address, cfg.CertFile, cfg.KeyFile)
	if err != nil {
		log.Fatal("failed to start repl client")
	}

	data := repl.ClientMsg{
		Id:         rc.State.Id,
		Head:       rc.State.Head,
		AuthSecret: cfg.AuthSecret,
	}
	dataEnc, _ := data.Encode()
	msg := protocol.Msg{
		Body: dataEnc,
	}
	msg.WriteTo(conn)

	log.Println("authed with repl master!! woohoooo :)")
	go rc.processOps(conn)
}

// ensures that repl file exists
func (rc *ReplClient) ensureStateFile() {
	replFilePath := path.Join(rc.Dir, REPL_FILENAME)
	replFile, err := os.Open(replFilePath)
	if err != nil {
		log.Println("no repl file found, will create...")
		// make new file
		newReplFile, err := os.Create(replFilePath)
		if err != nil {
			log.Fatalf("failed to create repl file, check directory exists: %s", rc.Dir)
		}

		// generate new id
		rc.State = &ReplClientState{}
		rc.State.Id = make([]byte, 32)
		rand.Seed(time.Now().UnixMicro())
		rand.Read(rc.State.Id)

		gob.NewEncoder(newReplFile).Encode(rc.State)
	} else {
		// decode list
		state := ReplClientState{}
		err := gob.NewDecoder(replFile).Decode(&state)
		if err != nil {
			log.Fatalf("failed to decode repl file: %s", err)
		}
		rc.State = &state
	}
	log.Printf("initialised repl client with id %s and head %v\r\n",
		base64.RawStdEncoding.EncodeToString(rc.State.Id), rc.State.Head)
}

func (rc *ReplClient) writeStateToFilePeriodically() {
	for {
		if rc.HeadIncremented {
			rc.writeStateToFile()
		}
		time.Sleep(time.Duration(10) * time.Second)
	}
}

func (rc *ReplClient) writeStateToFile() {
	fullPath := path.Join(rc.Dir, REPL_FILENAME)
	file, err := os.Create(fullPath)
	if err != nil {
		log.Fatalf("failed to create repl file: %s\r\n", err)
	}
	gob.NewEncoder(file).Encode(&rc.State)
	file.Close()
	rc.HeadIncremented = false
}

func (rc *ReplClient) processOps(conn net.Conn) {
	backoff := BACKOFF
	for {
		msg, err := protocol.ReadMsgFrom(conn)
		if err != nil {
			log.Println("failed to read repl op msg:", err)
			if backoff > BACKOFF_LIMIT {
				break
			}
			time.Sleep(time.Duration(backoff) * time.Millisecond)
			backoff *= 2
			continue
		}

		op := repl.Op{}
		err = op.DecodeFrom(msg.Body)
		if err != nil {
			log.Println("failed to decode repl op:", err)
			continue
		}

		log.Println("handle repl op:", op)
	}
	conn.Close()
}
