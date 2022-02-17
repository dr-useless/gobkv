package protocol

type StatusReply struct {
	Status byte
}

type ValueReply struct {
	Value   []byte
	Expires int64
}

type KeysReply struct {
	Keys []string
}

type Args struct {
	AuthSecret string
	Key        string
	Value      []byte
	Limit      int
	Expires    int64
}
