package protocol

const (
	OpClose  byte = 0x01 // close connection
	OpAuth   byte = 0x02 // authenticate
	OpPing   byte = 0x10 // ping server, responds with pong
	OpPong   byte = 0x11 // response to ping
	OpGet    byte = 0x20 // get value for given key
	OpSet    byte = 0x30 // set value of given key
	OpSetAck byte = 0x31 // set with OK response
	OpDel    byte = 0x40 // delete given key
	OpDelAck byte = 0x41 // delete with OK response
	OpList   byte = 0x50 // stream list of keys with prefix
	OpCount  byte = 0x60 // count keys with prefix
)

// Map of string labels for op codes
type Label map[byte]string

// Maps op codes to string labels
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
	}
}
