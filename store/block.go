package store

import (
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"os"
	"path"
	"sync"

	"github.com/intob/rocketkv/util"
)

// Second layer of division of the Store
//
// Child of Part
//
// Contains the slots with values and metadata
//
// MustWrite flag is true if changes have been made since last disk-write.
// MustSync flag is true for each node if changes have been made since last sync
type Block struct {
	Id        []byte
	Mutex     *sync.RWMutex
	Slots     map[string]Slot
	MustWrite bool
	ReplState map[uint64]*ReplNodeState // replNodeId
}

type ReplNodeState struct {
	MustSync bool
}

// Contains a value & associated metadata
type Slot struct {
	Value    []byte
	Expires  int64
	Modified int64
}

func NewBlock(id []byte) *Block {
	return &Block{
		Id:        id,
		Mutex:     new(sync.RWMutex),
		Slots:     make(map[string]Slot),
		ReplState: make(map[uint64]*ReplNodeState),
	}
}

// Returns checksum of all slot values
// For now, only sum Value (not Expires)
func (b *Block) Checksum() []byte {
	h := fnv.New128a()
	b.Mutex.RLock()
	for _, slot := range b.Slots {
		h.Write(slot.Value)
	}
	b.Mutex.RUnlock()
	return h.Sum(nil)
}

// Encodes block slots as gob, and writes to file
func (b *Block) WriteToFile(dir string) {
	if !b.MustWrite {
		return
	}
	name := util.GetName(b.Id)
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

// Decodes block file & populates slots
func (b *Block) ReadFromFile(dir string) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	name := util.GetName(b.Id)
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
