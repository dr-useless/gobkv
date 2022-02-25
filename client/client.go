package client

import (
	"errors"
	"net"

	"github.com/intob/chamux"
	"github.com/intob/gobkv/protocol"
)

const ErrEmptySecret = "secret is empty"
const ErrNegativeExpiry = "expires should be 0 or positive"
const ErrEmptyKey = "key must not be empty"

const MSG = "msg"

type Client struct {
	mconn   chamux.MConn
	msg     chamux.Topic
	msgSub  <-chan []byte
	MsgChan chan protocol.Msg
}

// Returns a pointer to a new Client,
// with topics, subscriptions & an output channel
// made ready.
// Messages can be receieved on MsgChan.
func NewClient(conn net.Conn) *Client {
	mc := chamux.NewMConn(conn, chamux.Gob{}, chamux.Options{})
	c := &Client{
		mconn: mc,
	}
	c.msg = chamux.NewTopic(MSG)
	c.msgSub = c.msg.Subscribe()
	c.mconn.AddTopic(&c.msg)
	c.MsgChan = make(chan protocol.Msg, 1)
	go c.pumpMsgs()
	return c
}

// Decodes & writes messages to output chan
func (c *Client) pumpMsgs() {
	for mBytes := range c.msgSub {
		msg, err := protocol.DecodeMsg(mBytes)
		if err != nil {
			panic(err)
		}
		c.MsgChan <- *msg
	}
}

// Encode & publish the given message
func (c *Client) EncodeAndPublish(msg *protocol.Msg) error {
	msgEnc, err := protocol.EncodeMsg(msg)
	if err != nil {
		return err
	}
	frame := chamux.NewFrame(msgEnc, MSG)
	return c.mconn.Publish(frame)
}

// Publish to the underlying MConn
func (c *Client) Publish(f *chamux.Frame) error {
	return c.mconn.Publish(f)
}

// Sends a ping message
//
// A status message will follow
func (c *Client) Ping() error {
	return c.EncodeAndPublish(&protocol.Msg{
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
	return c.EncodeAndPublish(msg)
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
	return c.EncodeAndPublish(msg)
}

// Get the value & expires time for a key
//
// The response will follow on the MsgChan
func (c *Client) Get(key string) error {
	msg := &protocol.Msg{
		Op:  protocol.OpGet,
		Key: key,
	}
	return c.EncodeAndPublish(msg)
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
	return c.EncodeAndPublish(msg)
}

// List all keys with the given prefix
func (c *Client) List(key string) error {
	msg := &protocol.Msg{
		Op:  protocol.OpList,
		Key: key,
	}
	return c.EncodeAndPublish(msg)
}
