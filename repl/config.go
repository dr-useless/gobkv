package repl

type ReplConfig struct {
	Name       string
	Id         []byte // calulated by hashing name
	Network    string
	Address    string
	AuthSecret string // for now, this is the same for all nodes
	CertFile   string
	KeyFile    string
	BufferSize int
	Peers      []PeerCfg // TODO: add service discovery
	Period     int       // seconds
}

type PeerCfg struct {
	Name    string
	Id      []byte // calculated by hashing name
	Network string
	Address string
}
