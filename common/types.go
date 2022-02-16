package common

type StatusReply struct {
	Status byte
}

type ValueReply struct {
	Value   []byte
	Expires uint32
}

type KeysReply struct {
	Keys []string
}

type Args struct {
	AuthSecret string
	Key        string
	Value      []byte
	Limit      int
	Expires    uint32
}
