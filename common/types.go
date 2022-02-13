package common

type Result struct {
	Status byte
	Value  []byte
	Keys   []string
}

type Args struct {
	AuthSecret string
	Key        string
	Value      []byte
}
