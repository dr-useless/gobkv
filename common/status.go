package common

const (
	StatusOk = '_'
)

func MapStatus() map[byte]string {
	m := make(map[byte]string)
	m[StatusOk] = "OK"
	return m
}
