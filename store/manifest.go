package store

import (
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/intob/gobkv/protocol"
)

type PartConfig struct {
	Count       int
	Persist     bool
	WritePeriod int // seconds
}

// ensures that a manifest & block files exist
func (s *Store) EnsureManifest(cfg *PartConfig) {
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
			partId := make([]byte, KEY_HASH_LEN)
			rand.Read(partId)
			part := Part{
				Id:     partId,
				Blocks: make(map[uint64]*Block),
			}
			// blocks
			for b := 0; b < cfg.Count; b++ {
				blockId := make([]byte, KEY_HASH_LEN)
				rand.Read(blockId)
				part.Blocks[getNumber(blockId)] = &Block{
					Id:    blockId,
					Mutex: new(sync.RWMutex),
					Slots: make(map[string]Slot),
				}
			}
			s.Parts[getNumber(partId)] = &part
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
					Slots: make(map[string]Slot),
				}
			}
			s.Parts[getNumber(part.Id)] = &part
		}
		blockCount := len(s.Parts) * len(s.Parts)
		fmt.Printf("initialised %v blocks from manifest\r\n", blockCount)
	}
}

func (s *Store) getManifest() *protocol.Manifest {
	manifest := make(protocol.Manifest, 0)
	for _, part := range s.Parts {
		partManifest := protocol.PartManifest{
			PartId: part.Id,
			Blocks: make([]protocol.BlockManifest, 0),
		}
		for _, block := range part.Blocks {
			blockManifest := protocol.BlockManifest{
				BlockId: block.Id,
				Hash:    block.Checksum(),
			}
			partManifest.Blocks = append(partManifest.Blocks, blockManifest)
		}
		manifest = append(manifest, partManifest)
	}
	return &manifest
}
