package protocol

const (
	OpPing byte = 0x10
	OpGet  byte = 0x20
	OpSet  byte = 0x30
	OpDel  byte = 0x40
	OpList byte = 0x50
)

func MapOp() map[byte]string {
	m := make(map[byte]string)
	m[OpPing] = "PING"
	m[OpGet] = "GET"
	m[OpSet] = "SET"
	m[OpDel] = "DEL"
	m[OpList] = "LIST"
	return m
}
