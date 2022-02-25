package protocol

const (
	StatusOk           byte = '_'
	StatusStreamEnd    byte = '/'
	StatusNotFound     byte = '.'
	StatusError        byte = '!'
	StatusUnauthorized byte = '#'
)

func MapStatus() Label {
	return Label{
		StatusOk:           "OK",
		StatusStreamEnd:    "STREAM_END",
		StatusNotFound:     "NOT_FOUND",
		StatusError:        "ERROR",
		StatusUnauthorized: "UNATHORIZED",
	}
}
