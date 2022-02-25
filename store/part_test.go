package store

import (
	"fmt"
	"testing"
)

func getTestPart(blocks int) Part {
	p := Part{
		Blocks: make(map[uint64]*Block),
	}
	for i := 0; i < blocks; i++ {
		// get random ID by hashing i
		id := hashKey(fmt.Sprintf("%v", i))
		p.Blocks[getNumber(id)] = &Block{
			Id: id,
		}
	}
	return p
}

func BenchmarkGetClosestBlock(b *testing.B) {
	keyHash := hashKey("test")
	part := getTestPart(256)
	for n := 0; n < b.N; n++ {
		part.getClosestBlock(keyHash)
	}
}
