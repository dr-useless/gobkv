package store

import (
	"bytes"
	"strings"

	"github.com/intob/rocketkv/util"
)

const manifestFileName = "manifest.gob"

// First layer of division of the Store
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
		util.FastXor(dist, keyHash, block.Id)
		if clDist == nil || bytes.Compare(dist, clDist) < 0 {
			clBlock = block
			clDist = dist
		}
		// reset dist
		dist = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	}

	return clBlock
}

// Lists all keys matching given prefix
//
// If a namespace is provided, results will be prefixed with it
func (p *Part) listKeys(namespace, prefix string, o chan string) {
	for _, block := range p.Blocks {
		block.Mutex.RLock()
		for k := range block.Slots {
			if strings.HasPrefix(k, prefix) {
				o <- namespace + k
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
