package store

import (
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"os"
	"path"

	"github.com/intob/rocketkv/cfg"
	"github.com/intob/rocketkv/util"
	"github.com/spf13/viper"
)

// Array of PartManifest
type Manifest []PartManifest

// Contains a PartId & array of blocks
type PartManifest struct {
	PartId []byte
	Blocks []BlockManifest
}

// Contains a BlockId
type BlockManifest struct {
	BlockId []byte
}

// ensureManifest ensures that a manifest & block files exist
func ensureManifest(s *Store) {
	if !viper.GetBool(cfg.PERSIST) {
		return
	}
	s.Parts = make(map[uint64]*Part)
	manifestPath := path.Join(s.Dir, manifestFileName)
	manifestFile, err := os.Open(manifestPath)
	segments := viper.GetInt(cfg.SEGMENTS)
	if err != nil {
		fmt.Println("no manifest found, will create...")
		// parts
		for p := 0; p < segments; p++ {
			partId := make([]byte, util.ID_LEN)
			_, err := rand.Read(partId)
			if err != nil {
				fmt.Println("failed to read from rand reader")
				panic(err)
			}
			part := NewPart(partId)
			// blocks
			for b := 0; b < segments; b++ {
				blockId := make([]byte, util.ID_LEN)
				rand.Read(blockId)
				part.Blocks[util.GetNumber(blockId)] = NewBlock(blockId)
			}
			s.Parts[util.GetNumber(partId)] = &part
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
		manifest := make(Manifest, 0)
		err := gob.NewDecoder(manifestFile).Decode(&manifest)
		if err != nil {
			fmt.Println("failed to decode manifest")
			panic(err)
		}
		for _, partManifest := range manifest {
			part := NewPart(partManifest.PartId)
			for _, block := range partManifest.Blocks {
				part.Blocks[util.GetNumber(block.BlockId)] = NewBlock(block.BlockId)
			}
			s.Parts[util.GetNumber(part.Id)] = &part
		}
		blockCount := len(s.Parts) * len(s.Parts)
		fmt.Printf("initialised %v blocks from manifest\r\n", blockCount)
	}
}

// getManifest returns a pointer to a new manifest
func (s *Store) getManifest() *Manifest {
	manifest := make(Manifest, 0)
	for _, part := range s.Parts {
		partManifest := PartManifest{
			PartId: part.Id,
			Blocks: make([]BlockManifest, 0),
		}
		for _, block := range part.Blocks {
			blockManifest := BlockManifest{
				BlockId: block.Id,
			}
			partManifest.Blocks = append(partManifest.Blocks, blockManifest)
		}
		manifest = append(manifest, partManifest)
	}
	return &manifest
}
