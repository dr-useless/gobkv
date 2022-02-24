package client

import (
	"errors"
	"net"

	"github.com/intob/chamux"
	"github.com/intob/gobkv/protocol"
)

const PF = "gobkv: "
const MSG = "msg"

type Client struct {
	mconn   chamux.MConn
	msg     chamux.Topic
	msgSub  <-chan []byte
	MsgChan chan protocol.Msg
}

func NewClient(conn net.Conn) *Client {
	mc := chamux.NewMConn(conn, chamux.Gob{}, 2048)
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

func (c *Client) pumpMsgs() {
	for mBytes := range c.msgSub {
		msg, err := protocol.DecodeMsg(mBytes)
		if err != nil {
			panic(err)
		}
		c.MsgChan <- *msg
	}
}

func (c *Client) Ping() error {
	msgEnc, err := protocol.EncodeMsg(&protocol.Msg{
		Op: protocol.OpPing,
	})
	if err != nil {
		return err
	}
	frame := chamux.NewFrame(msgEnc, MSG)
	return c.mconn.Publish(frame)
}

func (c *Client) Auth(secret string) error {
	if secret == "" {
		return errors.New(PF + "secret is empty")
	}
	msgEnc, err := protocol.EncodeMsg(&protocol.Msg{
		Op:  protocol.OpAuth,
		Key: secret,
	})
	if err != nil {
		return err
	}
	frame := chamux.NewFrame(msgEnc, MSG)
	return c.mconn.Publish(frame)
}
