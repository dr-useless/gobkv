package client

import (
	"bufio"
	"errors"
	"net"

	"github.com/intob/rocketkv/protocol"
)

const ErrEmptySecret = "secret is empty"
const ErrNegativeExpiry = "expires should be 0 or positive"
const ErrEmptyKey = "key must not be empty"

const MSG = "msg"

type Client struct {
	conn net.Conn
	Msgs chan protocol.Msg
}

// Returns a pointer to a new Client,
// Messages can be receieved on Msgs chan.
func NewClient(conn net.Conn) *Client {
	c := &Client{
		conn: conn,
		Msgs: make(chan protocol.Msg),
	}
	go c.pumpMsgs()
	return c
}

// Reads from conn, decodes & writes
// messages to Msgs chan
func (c *Client) pumpMsgs() {
	scan := bufio.NewScanner(c.conn)
	scan.Split(protocol.SplitPlusEnd)
	for scan.Scan() {
		mBytes := scan.Bytes()
		msg, err := protocol.DecodeMsg(mBytes)
		if err != nil {
			panic(err)
		}
		c.Msgs <- *msg
	}
}

// Encode & publish the given message
func (c *Client) Send(msg *protocol.Msg) error {
	msgEnc, err := protocol.EncodeMsg(msg)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(msgEnc)
	return err
}

// Sends a ping message
//
// A status message will follow
func (c *Client) Ping() error {
	return c.Send(&protocol.Msg{
		Op: protocol.OpPing,
	})
}

// Authenticate using the given secret
//
// A status message will follow
func (c *Client) Auth(secret string) error {
	if secret == "" {
		return errors.New(ErrEmptySecret)
	}
	msg := &protocol.Msg{
		Op:  protocol.OpAuth,
		Key: secret,
	}
	return c.Send(msg)
}

// Sets the value & expires properties of the key
//
// If expires is 0, the key will not expire
// If ack is true, a response will follow
func (c *Client) Set(key string, value []byte, expires int64, ack bool) error {
	if expires < 0 {
		return errors.New(ErrNegativeExpiry)
	}
	if key == "" {
		return errors.New(ErrEmptyKey)
	}
	msg := &protocol.Msg{
		Op:    protocol.OpSet,
		Key:   key,
		Value: value,
	}
	if ack {
		msg.Op = protocol.OpSetAck
	}
	return c.Send(msg)
}

// Get the value & expires time for a key
//
// The response will follow on the MsgChan
func (c *Client) Get(key string) error {
	msg := &protocol.Msg{
		Op:  protocol.OpGet,
		Key: key,
	}
	return c.Send(msg)
}

// Delete a key
//
// If ack is true, a status response will follow
func (c *Client) Del(key string, ack bool) error {
	msg := &protocol.Msg{
		Op:  protocol.OpDel,
		Key: key,
	}
	if ack {
		msg.Op = protocol.OpDelAck
	}
	return c.Send(msg)
}

// List all keys with the given prefix
func (c *Client) List(key string) error {
	msg := &protocol.Msg{
		Op:  protocol.OpList,
		Key: key,
	}
	return c.Send(msg)
}
