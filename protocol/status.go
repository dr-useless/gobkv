package protocol

const (
	StatusError byte = '!'
	StatusOk    byte = '_'
)

func MapStatus() map[byte]string {
	m := make(map[byte]string)
	m[StatusOk] = "OK"
	m[StatusError] = "ERROR"
	return m
}
