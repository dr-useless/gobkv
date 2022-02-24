package store

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"os"
	"path"
	"sync"

	"github.com/intob/gobkv/protocol"
)

const manifestFileName = "manifest.gob"
const idLen = 8

type Part struct {
	Id     []byte
	Blocks map[uint64]*Block
}

type Block struct {
	Id        []byte
	Mutex     *sync.RWMutex
	Slots     map[string]*Slot
	MustWrite bool
}

type Slot struct {
	Value    []byte
	Expires  int64
	Modified int64
}

type PartConfig struct {
	Count       int
	Persist     bool
	WritePeriod int // seconds
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
	gob.NewEncoder(file).Encode(&b.Slots)
	b.MustWrite = false
	b.Mutex.RUnlock()
	file.Close()
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

// ensures that part files exist
func (s *Store) EnsureBlocks(cfg *PartConfig) {
	if !cfg.Persist {
		return
	}
	s.Parts = make(map[uint64]*Part)
	manifestPath := path.Join(s.Dir, manifestFileName)
	manifestFile, err := os.Open(manifestPath)
	if err != nil {
		fmt.Println("no manifest found, will create...")
		// parts
		for p := 0; p < cfg.Count; p++ {
			partId := make([]byte, idLen)
			rand.Read(partId)
			part := &Part{
				Id:     partId,
				Blocks: make(map[uint64]*Block),
			}
			// blocks
			for b := 0; b < cfg.Count; b++ {
				blockId := make([]byte, idLen)
				rand.Read(blockId)
				part.Blocks[getNumber(blockId)] = &Block{
					Id:    blockId,
					Mutex: new(sync.RWMutex),
					Slots: make(map[string]*Slot),
				}
			}
			s.Parts[getNumber(partId)] = part
		}
		newManifestFile, err := os.Create(manifestPath)
		if err != nil {
			fmt.Printf("failed to create manifest, check directory exists: %s\r\n", s.Dir)
			panic(err)
		}
		manifest := s.getManifest()
		gob.NewEncoder(newManifestFile).Encode(manifest)
	} else {
		// decode list
		manifest := make(protocol.Manifest, 0)
		err := gob.NewDecoder(manifestFile).Decode(&manifest)
		if err != nil {
			fmt.Println("failed to decode manifest")
			panic(err)
		}
		for _, partManifest := range manifest {
			part := Part{
				Id:     partManifest.PartId,
				Blocks: make(map[uint64]*Block),
			}
			for _, block := range partManifest.Blocks {
				part.Blocks[getNumber(block.BlockId)] = &Block{
					Id:    block.BlockId,
					Mutex: new(sync.RWMutex),
					Slots: make(map[string]*Slot),
				}
			}
			s.Parts[getNumber(part.Id)] = &part
		}
		blockCount := len(s.Parts) * len(s.Parts)
		fmt.Printf("initialised %v blocks from manifest\r\n", blockCount)
	}
}

func (p *Part) getClosestBlock(key string) *Block {
	h := fnv.New64a()
	h.Write([]byte(key))
	keyHash := h.Sum(nil)
	var clBlock *Block
	var clD []byte
	for _, block := range p.Blocks {
		d := xorBytes(block.Id, keyHash)
		if clD == nil || bytes.Compare(d, clD) < 0 {
			clBlock = block
			clD = d
		}
	}
	return clBlock
}

func getName(id []byte) string {
	return base64.RawURLEncoding.EncodeToString(id)
}

/*func getId(name string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(name)
}*/

func getNumber(id []byte) uint64 {
	return binary.BigEndian.Uint64(id)
}
