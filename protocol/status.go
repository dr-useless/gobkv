package protocol

const (
	StatusOk    byte = '_'
	StatusError byte = '!'
)

func MapStatus() Label {
	return Label{
		StatusOk:    "OK",
		StatusError: "ERROR",
	}
}
