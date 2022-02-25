package store

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"

	"github.com/lukechampine/fastxor"
)

const manifestFileName = "manifest.gob"
const idLen = 8

type Part struct {
	Id     []byte
	Blocks map[uint64]*Block
}

func (p *Part) getClosestBlock(keyHash []byte) *Block {
	var clDist []byte  // smallest distance value seen
	var clBlock *Block // block with smallest distance
	dist := make([]byte, KEY_HASH_LEN)

	// range through blocks to find closest
	for _, block := range p.Blocks {
		fastxor.Bytes(dist, keyHash, block.Id)
		if clDist == nil || bytes.Compare(dist, clDist) < 0 {
			clBlock = block
			clDist = dist
		}
	}

	return clBlock
}

func getName(id []byte) string {
	return base64.RawURLEncoding.EncodeToString(id)
}

func getNumber(id []byte) uint64 {
	return binary.BigEndian.Uint64(id)
}
