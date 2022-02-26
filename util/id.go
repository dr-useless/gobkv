package util

import "crypto/rand"

func RandomId() ([]byte, error) {
	id := make([]byte, ID_LEN)
	_, err := rand.Read(id)
	return id, err
}
