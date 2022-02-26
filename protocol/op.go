package protocol

const (
	OpClose  byte = 0x01
	OpAuth   byte = 0x02
	OpPing   byte = 0x10
	OpPong   byte = 0x11
	OpGet    byte = 0x20
	OpSet    byte = 0x30
	OpSetAck byte = 0x31
	OpDel    byte = 0x40
	OpDelAck byte = 0x41
	OpList   byte = 0x50
	OpCount  byte = 0x60
	OpBlock  byte = 0xA0
)

type Label map[byte]string

func MapOp() Label {
	return Label{
		OpClose:  "CLOSE",
		OpAuth:   "AUTH",
		OpPing:   "PING",
		OpPong:   "PONG",
		OpGet:    "GET",
		OpSet:    "SET",
		OpSetAck: "SET_ACK",
		OpDel:    "DEL",
		OpDelAck: "DEL_ACK",
		OpList:   "LIST",
		OpBlock:  "BLOCK",
	}
}
