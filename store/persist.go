package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/intob/rocketkv/cfg"
	"github.com/spf13/viper"
)

// While watch() only takes care of writing to partitions,
// will only watch if persistence is enabled
func (st *Store) Persist(dir string, period int) {

	fmt.Printf("will write changed partitions every %v seconds\r\n", period)
	for {
		st.WriteAllBlocks(dir)
		time.Sleep(time.Duration(period) * time.Second)
	}
}

func (st *Store) WriteAllBlocks(dir string) {
	for _, part := range st.Parts {
		for _, block := range part.Blocks {
			go func(b *Block) {
				b.WriteToFile(dir)
			}(block)
		}
	}
}

func readFromBlockFiles(st *Store) {
	if !viper.GetBool(cfg.PERSIST) {
		return
	}
	dir := viper.GetString(cfg.DIR)
	wg := new(sync.WaitGroup)
	for _, part := range st.Parts {
		for _, b := range part.Blocks {
			wg.Add(1)
			go func(b *Block) {
				b.ReadFromFile(dir)
				wg.Done()
			}(b)
		}
	}
	wg.Wait()
}
