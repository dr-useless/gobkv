package common

type StatusReply struct {
	Status byte
}

type ValueReply struct {
	Value []byte
}

type KeysReply struct {
	Keys []string
}

type Args struct {
	AuthSecret string
	Key        string
	Value      []byte
	Limit      int
}
