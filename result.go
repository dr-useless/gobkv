package main

import "encoding/binary"

const (
	StatusOk    = '_'
	StatusError = '!'
	StatusWarn  = '?'
)

type Result struct {
	Status byte
	Value  []byte
}

// Marshals a result into the following format
// _ | _ _ _ _ | _...
// Status | Value Length | Value
func (r *Result) Marshal() []byte {
	if len(r.Value) == 0 {
		// no value, return only status
		out := make([]byte, 1)
		out[0] = r.Status
		return out
	}

	// + 1 for status
	// + 4 for value length
	out := make([]byte, len(r.Value)+5)

	// copy status
	out[0] = r.Status

	// copy value length
	lenValue := make([]byte, 4)
	binary.BigEndian.PutUint32(lenValue, uint32(len(r.Value)))
	copy(out[1:], lenValue)

	// copy value
	copy(out[5:], r.Value)

	return out
}
