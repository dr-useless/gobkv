package store

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/intob/rocketkv/util"
)

func getTestStore(parts int, putSlots bool) *Store {
	s := &Store{
		Parts: make(map[uint64]*Part),
	}
	// make parts
	for i := 0; i < parts; i++ {
		part := getTestPart(parts, putSlots)
		s.Parts[util.GetNumber(part.Id)] = &part
	}
	return s
}

// Tests that calling getClosestBlock always returns
// the same block.
func TestGetClosestPart(t *testing.T) {
	keyHash := util.HashStr("test")
	store := getTestStore(8, false)
	clCtl := store.getClosestPart(keyHash)
	for i := 0; i < len(store.Parts); i++ {
		clCur := store.getClosestPart(keyHash)
		if !bytes.Equal(clCtl.Id, clCur.Id) {
			t.FailNow()
		}
	}
}

func TestSetAndGet(t *testing.T) {
	s := getTestStore(8, false)
	key := "test"
	value := []byte("coffee")

	s.Set(key, Slot{
		Value: value,
	}, false)

	got, found := s.Get(key)

	if !found {
		t.FailNow()
	}

	if !bytes.Equal(got.Value, value) {
		t.FailNow()
	}
}

func TestSetAndGetWithNamespace(t *testing.T) {
	s := getTestStore(8, false)
	key := "mynamespace/somecollection/test"
	value := []byte("coffee")

	s.Set(key, Slot{
		Value: value,
	}, false)

	got, found := s.Get(key)

	if !found {
		t.FailNow()
	}

	if !bytes.Equal(got.Value, value) {
		t.FailNow()
	}
}

func TestSetAndDel(t *testing.T) {
	s := getTestStore(8, false)
	key := "test"
	value := []byte("coffee")

	s.Set(key, Slot{
		Value: value,
	}, false)

	s.Del(key)

	_, found := s.Get(key)

	if found {
		t.FailNow()
	}
}

func TestSetAndDelWithNamespace(t *testing.T) {
	s := getTestStore(8, false)
	key := "mynamespace/collection/test"
	value := []byte("coffee")

	s.Set(key, Slot{
		Value: value,
	}, false)

	s.Del(key)

	_, found := s.Get(key)

	if found {
		t.FailNow()
	}
}

func TestList(t *testing.T) {
	s := getTestStore(8, false)

	count := 1000
	keyPrefix := "somekeyprefix_"
	slot := Slot{Value: []byte("test")}

	// populate store with test data
	for i := 0; i < count; i++ {
		key := keyPrefix + strconv.Itoa(i)
		s.Set(key, slot, false)
	}

	keyChan := s.List(keyPrefix, 1)

	// tally up keys, expecting count
	tally := 0
	for range keyChan {
		tally++
	}

	if tally != count {
		t.FailNow()
	}
}

func TestListWithNamespace(t *testing.T) {
	s := getTestStore(8, false)

	count := 1000
	keyPrefix := "namespace/collectionofthings/somekeyprefix_"
	slot := Slot{Value: []byte("test")}

	// populate store with test data
	for i := 0; i < count; i++ {
		key := keyPrefix + strconv.Itoa(i)
		s.Set(key, slot, false)
	}

	// tally up keys, expecting count
	keyChan := s.List(keyPrefix, 1)
	tally := 0
	for range keyChan {
		tally++
	}

	if tally != count {
		t.FailNow()
	}
}

func TestCount(t *testing.T) {
	s := getTestStore(8, false)

	count := 1000
	keyPrefix := "somekeyprefix_"
	slot := Slot{Value: []byte("test")}

	// populate store with test data
	for i := 0; i < count; i++ {
		key := keyPrefix + strconv.Itoa(i)
		s.Set(key, slot, false)
	}

	if uint64(count) != s.Count(keyPrefix) {
		t.FailNow()
	}
}

func TestCountWithNamespace(t *testing.T) {
	s := getTestStore(8, false)

	count := 1000
	keyPrefix := "mynamespace/collection/somekeyprefix_"
	slot := Slot{Value: []byte("test")}

	// populate store with test data
	for i := 0; i < count; i++ {
		key := keyPrefix + strconv.Itoa(i)
		s.Set(key, slot, false)
	}

	if uint64(count) != s.Count(keyPrefix) {
		t.FailNow()
	}
}
