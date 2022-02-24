package client

import (
	"net"

	"github.com/intob/chamux"
	"github.com/intob/gobkv/protocol"
)

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
	go c.pumpMsgs()
	return c
}

func (c *Client) pumpMsgs() {
	for msg := range c.msgSub {
		m := protocol.Msg{}
		err := m.DecodeFrom(msg)
		if err != nil {
			panic(err)
		}
		c.MsgChan <- m
	}
}

func (c *Client) Ping() {
	msg := protocol.Msg{
		Op: protocol.OpPing,
	}
	msgEnc, _ := msg.Encode()
	frame := chamux.NewFrame(msgEnc, MSG)
	c.mconn.Publish(frame)
}
