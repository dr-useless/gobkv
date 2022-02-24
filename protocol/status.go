package protocol

const (
	StatusOk           byte = '_'
	StatusError        byte = '!'
	StatusUnauthorized byte = '#'
	StatusNotFound     byte = '.'
)

func MapStatus() Label {
	return Label{
		StatusOk:           "OK",
		StatusError:        "ERROR",
		StatusUnauthorized: "UNATHORIZED",
		StatusNotFound:     "NOT_FOUND",
	}
}
