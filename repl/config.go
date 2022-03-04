package repl

type PeerCfg struct {
	NetAddr
	Name string
	Id   []byte // calculated by hashing name
}

type NetAddr struct {
	Network string
	Address string
}
