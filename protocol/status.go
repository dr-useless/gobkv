package protocol

const (
	StatusOk           byte = '_'
	StatusError        byte = '!'
	StatusUnauthorized byte = '/'
	StatusNotFound     byte = '0'
)

func MapStatus() Label {
	return Label{
		StatusOk:           "OK",
		StatusError:        "ERROR",
		StatusUnauthorized: "UNAUTHORIZED",
		StatusNotFound:     "NOT_FOUND",
	}
}
