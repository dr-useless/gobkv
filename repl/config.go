package repl

type ReplConfig struct {
	Network  string
	Address  string
	CertFile string
	KeyFile  string
	Peers    []ReplPeer
	id       []byte // randomly generated on launch
}

type ReplPeer struct {
	Id      []byte
	Network string
	Address string
}
