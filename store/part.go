package store

import (
	"bytes"

	"github.com/lukechampine/fastxor"
)

const manifestFileName = "manifest.gob"

type Part struct {
	Id     []byte
	Blocks map[uint64]*Block
}

// Returns pointer to block with least Hamming distance
// from given key hash
func (p *Part) getClosestBlock(keyHash []byte) *Block {
	var clDist []byte  // winning distance
	var clBlock *Block // winning block
	dist := make([]byte, KEY_HASH_LEN)

	// range through blocks to find closest
	for _, block := range p.Blocks {
		fastxor.Block(dist, keyHash, block.Id)
		if clDist == nil || bytes.Compare(dist, clDist) < 0 {
			clBlock = block
			clDist = dist
		}
	}

	return clBlock
}
