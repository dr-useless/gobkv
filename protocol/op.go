package protocol

const (
	OpPing byte = 0x10
	OpPong byte = 0x11
	OpGet  byte = 0x20
	OpSet  byte = 0x30
	OpDel  byte = 0x40
	OpList byte = 0x50
)

func MapOp() map[byte]string {
	m := make(map[byte]string)
	m[OpPing] = "PING"
	m[OpPong] = "PONG"
	m[OpGet] = "GET"
	m[OpSet] = "SET"
	m[OpDel] = "DEL"
	m[OpList] = "LIST"
	return m
}
