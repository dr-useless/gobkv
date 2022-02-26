package store

import (
	"bytes"
	"testing"

	"github.com/intob/rocketkv/util"
)

func getTestStore(parts int) *Store {
	s := &Store{
		Parts: make(map[uint64]*Part),
	}
	// make parts
	for i := 0; i < parts; i++ {
		part := getTestPart(parts)
		s.Parts[getNumber(part.Id)] = &part
	}
	return s
}

// Tests that calling getClosestBlock always returns
// the same block.
func TestGetClosestPart(t *testing.T) {
	keyHash := util.HashKey("test")
	store := getTestStore(16)
	clCtl := store.getClosestPart(keyHash)
	for i := 0; i < len(store.Parts); i++ {
		clCur := store.getClosestPart(keyHash)
		if !bytes.Equal(clCtl.Id, clCur.Id) {
			t.FailNow()
		}
	}

}
