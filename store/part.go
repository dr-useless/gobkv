package store

import (
	"bytes"
	"strings"

	"github.com/intob/rocketkv/util"
	"github.com/lukechampine/fastxor"
)

const manifestFileName = "manifest.gob"

type Part struct {
	Id     []byte
	Blocks map[uint64]*Block
}

func NewPart(id []byte) Part {
	return Part{
		Id:     id,
		Blocks: make(map[uint64]*Block),
	}
}

// Returns pointer to block with least Hamming distance
// from given key hash
func (p *Part) getClosestBlock(keyHash []byte) *Block {
	var clDist []byte  // winning distance
	var clBlock *Block // winning block
	dist := make([]byte, util.ID_LEN)

	// range through blocks to find closest
	for _, block := range p.Blocks {
		fastxor.Block(dist, keyHash, block.Id)
		if clDist == nil || bytes.Compare(dist, clDist) < 0 {
			clBlock = block
			clDist = dist
		}
		// reset dist
		dist = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	}

	return clBlock
}

func (p *Part) listKeys(prefix string, o chan string) {
	for _, block := range p.Blocks {
		block.Mutex.RLock()
		for k := range block.Slots {
			if strings.HasPrefix(k, prefix) {
				o <- k
			}
		}
		block.Mutex.RUnlock()
	}
}

func (p *Part) countKeys(prefix string) uint64 {
	var count uint64
	for _, block := range p.Blocks {
		block.Mutex.RLock()
		for k := range block.Slots {
			if strings.HasPrefix(k, prefix) {
				count++
			}
		}
		block.Mutex.RUnlock()
	}
	return count
}
