package main

import (
	"bufio"
	"io"
	"net/textproto"
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

func (r *Result) Write(w io.Writer) (int, error) {
	if len(r.Value) == 0 {
		// no value, return only status
		return w.Write([]byte{r.Status, '\r', '\n', '.', '\r', '\n'})
	}

	bw := bufio.NewWriter(w)

	// status
	err := bw.WriteByte(r.Status)
	if err != nil {
		return 1, err
	}

	// value
	dotW := textproto.NewWriter(bw).DotWriter()
	defer dotW.Close()
	defer bw.Flush()
	return dotW.Write(r.Value)
}
