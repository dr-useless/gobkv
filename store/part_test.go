package store

import (
	"math/rand"
	"testing"

	"github.com/intob/gobkv/util"
)

func getTestPart(blocks int) Part {
	p := Part{
		Blocks: make(map[uint64]*Block),
	}
	for i := 0; i < blocks; i++ {
		// make block with random ID
		id := make([]byte, util.ID_LEN)
		rand.Read(id)
		b := NewBlock(id)
		// fill with blocks*256 random slots
		slotId := make([]byte, util.ID_LEN)
		for s := 0; s < blocks*256; s++ {
			_, err := rand.Read(slotId)
			if err != nil {
				panic(err)
			}
			b.Slots[getName(slotId)] = Slot{Value: slotId}
		}
		p.Blocks[getNumber(id)] = b
	}
	return p
}

var keyHash = util.HashKey("test")
var part = getTestPart(16)

func BenchmarkGetClosestBlock(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		part.getClosestBlock(keyHash)
	}
}

func BenchmarkListKeys(b *testing.B) {
	out := make(chan string)
	go func() {
		for {
			<-out
		}
	}()
	for i := 0; i < b.N; i++ {
		part.listKeys("", out)
	}
}
