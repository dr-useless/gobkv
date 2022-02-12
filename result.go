package main

import (
	"bufio"
	"io"
	"log"
)

const (
	StatusOk    = '_'
	StatusError = '!'
	StatusWarn  = '?'
)

type Result struct {
	Status byte
	Value  []byte
}

func (r *Result) WriteOn(w io.Writer) (int, error) {
	if len(r.Value) == 0 {
		// no value, return only status
		return w.Write([]byte{r.Status, '\n'})
	}

	bw := bufio.NewWriter(w)

	// status
	err := bw.WriteByte(r.Status)
	if err != nil {
		return 1, err
	}

	// value
	len, err := bw.Write(r.Value)
	if err != nil {
		log.Println("write error ", err)
		return len + 1, err
	}
	bw.WriteByte('\n')
	bw.Flush()
	return len + 1, nil
}
