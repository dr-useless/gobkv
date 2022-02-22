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
	State     *ReplClientState
	MustWrite bool
	Store     *Store
	Dir       string
}

type ReplClientState struct {
	ClientId []byte
	ReplId   []byte
	Head     int
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
		log.Fatalf("failed to connect to master at %s over %s", cfg.Address, cfg.Network)
	}

	body := repl.ClientMsgBody{
		ClientId:   rc.State.ClientId,
		AuthSecret: cfg.AuthSecret,
		ReplId:     rc.State.ReplId,
		Head:       rc.State.Head,
	}
	bodyEnc, _ := body.Encode()
	msg := protocol.Msg{
		Body: bodyEnc,
	}
	msg.WriteTo(conn)

	resp, err := protocol.ReadMsgFrom(conn)
	if err != nil {
		log.Fatal(err)
	}
	masterMsgBody := repl.MasterMsgBody{}
	masterMsgBody.DecodeFrom(resp.Body)

	rc.State.ReplId = masterMsgBody.ReplId
	rc.State.Head = masterMsgBody.Head

	rc.writeStateToFile()

	if masterMsgBody.MustSync {
		// fully sync
		log.Println("repl must fully resync")
	} else {
		log.Println("repl will resume")
	}

	log.Printf("\r\nrepl id: %s\r\nhead: %v\r\nclient id: %s",
		base64.RawStdEncoding.EncodeToString(rc.State.ReplId),
		rc.State.Head,
		base64.RawStdEncoding.EncodeToString(rc.State.ClientId))

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
		rc.State.ClientId = make([]byte, 32)
		rand.Seed(time.Now().UnixMicro())
		rand.Read(rc.State.ClientId)

		gob.NewEncoder(newReplFile).Encode(rc.State)
	} else {
		state := ReplClientState{}
		err := gob.NewDecoder(replFile).Decode(&state)
		if err != nil {
			log.Fatalf("failed to decode repl file: %s", err)
		}
		rc.State = &state
	}
}

func (rc *ReplClient) writeStateToFilePeriodically() {
	for {
		if rc.MustWrite {
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
	rc.MustWrite = false
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

		switch op.Op {
		case protocol.OpSet:
			rc.Store.Set(op.Key, &Slot{
				Value:    op.Value,
				Expires:  op.Expires,
				Modified: op.Modified,
			})
		case protocol.OpDel:
			rc.Store.Del(op.Key)
		default:
			log.Println("cannot replicate op", protocol.MapOp()[op.Op])
		}

		log.Println("handled repl op:", op)
	}
	conn.Close()
}
