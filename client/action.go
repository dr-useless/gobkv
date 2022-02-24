package client

import (
	"errors"

	"github.com/intob/chamux"
	"github.com/intob/gobkv/protocol"
)

func (c *Client) set(key string, value []byte, expires int64, ack bool) error {
	if expires < 0 {
		return errors.New(PF + "expires should be 0 or positive")
	}
	if key == "" {
		return errors.New(PF + "key must not be empty")
	}
	msg := &protocol.Msg{
		Op:    protocol.OpSet,
		Key:   key,
		Value: value,
	}
	if ack {
		msg.Op = protocol.OpSetAck
	}
	msgEnc, err := protocol.EncodeMsg(msg)
	if err != nil {
		return err
	}
	frame := chamux.NewFrame(msgEnc, MSG)
	return c.mconn.Publish(frame)
}
