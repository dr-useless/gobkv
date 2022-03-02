package repl

type ReplConfig struct {
	PeerCfg
	AuthSecret string // for now, this is the same for all nodes
	CertFile   string
	KeyFile    string
	BufferSize int
	Peers      []PeerCfg // TODO: add service discovery
	Period     int       // seconds
}

type PeerCfg struct {
	NetAddr
	Name string
	Id   []byte // calculated by hashing name
}

type NetAddr struct {
	Network string
	Address string
}
