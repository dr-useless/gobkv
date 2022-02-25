package store

import (
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"os"
	"path"
	"sync"
)

type Block struct {
	Id        []byte
	Mutex     *sync.RWMutex
	Slots     map[string]Slot
	MustWrite bool
}

type Slot struct {
	Value    []byte
	Expires  int64
	Modified int64
}

// Returns checksum of all slot values
// For now, only sum Value (not Expires)
func (b *Block) Checksum() []byte {
	h := fnv.New128a()
	b.Mutex.RLock()
	for _, slot := range b.Slots {
		h.Write(slot.Value)
	}
	defer b.Mutex.RUnlock()
	return h.Sum(nil)
}

func (b *Block) WriteToFile(dir string) {
	if !b.MustWrite {
		return
	}
	name := getName(b.Id)
	fullPath := path.Join(dir, name+".gob")
	file, err := os.Create(fullPath)
	if err != nil {
		fmt.Println("failed to create part file")
		panic(err)
	}
	b.Mutex.RLock()
	err = gob.NewEncoder(file).Encode(&b.Slots)
	if err != nil {
		panic(err)
	}
	file.Close()
	b.MustWrite = false
	b.Mutex.RUnlock()
}

func (b *Block) ReadFromFile(wg *sync.WaitGroup, dir string) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	defer wg.Done()
	name := getName(b.Id)
	fullPath := path.Join(dir, name+".gob")
	file, err := os.Open(fullPath)
	if err != nil {
		return
	}
	err = gob.NewDecoder(file).Decode(&b.Slots)
	file.Close()
	if err != nil {
		fmt.Printf("failed to decode data in block %s\r\n", name)
		return
	}
	fmt.Printf("read from block %s\r\n", name)
}
