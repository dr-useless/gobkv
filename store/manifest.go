package store

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"os"
	"path"

	"github.com/intob/rocketkv/util"
)

type PartConfig struct {
	Count       int
	Persist     bool
	WritePeriod int // seconds
}

type Manifest []PartManifest

type PartManifest struct {
	PartId []byte
	Blocks []BlockManifest
}

type BlockManifest struct {
	BlockId []byte
}

func (v *Manifest) DecodeFrom(b []byte) error {
	var buf bytes.Buffer
	buf.Write(b)
	return gob.NewDecoder(&buf).Decode(v)
}

func (v *Manifest) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
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
			partId := make([]byte, util.ID_LEN)
			_, err := rand.Read(partId)
			if err != nil {
				fmt.Println("failed to read from rand reader")
				panic(err)
			}
			part := NewPart(partId)
			// blocks
			for b := 0; b < cfg.Count; b++ {
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
