package protocol

// Msg body for normal ops
type Data struct {
	Key     string
	Value   []byte
	Expires int64
	Keys    []string
}
