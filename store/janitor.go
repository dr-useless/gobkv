package store

import (
	"fmt"
	"sync"
	"time"
)

// Delete expired keys
func scanForExpiredKeys(s *Store, scanPeriod int) {
	fmt.Printf("will scan for expired keys every %v seconds\r\n", scanPeriod)
	for {
		wg := new(sync.WaitGroup)
		for _, part := range s.Parts {
			for _, block := range part.Blocks {
				wg.Add(1)
				go func(block *Block, dir string) {
					defer wg.Done()
					block.Mutex.RLock()
					for k, slot := range block.Slots {
						if slot.Expires == 0 {
							continue
						}
						expires := time.Unix(slot.Expires, 0)
						if time.Now().After(expires) {
							block.Mutex.RUnlock()
							block.Mutex.Lock()
							delete(block.Slots, k)
							block.MustWrite = true
							block.Mutex.Unlock()
							block.Mutex.RLock()
						}
					}
					block.Mutex.RUnlock()
				}(block, s.Dir)
			}
			time.Sleep(time.Duration(scanPeriod) * time.Second)
		}
		wg.Wait()
	}
}
