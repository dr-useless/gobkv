package protocol

const (
	OpClose  byte = 0x01
	OpPing   byte = 0x10
	OpPong   byte = 0x11
	OpGet    byte = 0x20
	OpSet    byte = 0x30
	OpSetAck byte = 0x31 // respond with status
	OpDel    byte = 0x40
	OpDelAck byte = 0x41 // respond with status
	OpList   byte = 0x50
)

func MapOp() map[byte]string {
	m := make(map[byte]string)
	m[OpClose] = "CLOSE"
	m[OpPing] = "PING"
	m[OpPong] = "PONG"
	m[OpGet] = "GET"
	m[OpSet] = "SET"
	m[OpSetAck] = "SET_ACK"
	m[OpDel] = "DEL"
	m[OpDelAck] = "DEL_ACK"
	m[OpList] = "LIST"
	return m
}
