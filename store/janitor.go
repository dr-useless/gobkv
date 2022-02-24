package store

import (
	"log"
	"sync"
	"time"
)

// Delete expired keys
func (s *Store) ScanForExpiredKeys(cfg *PartConfig, scanPeriod int) {
	if scanPeriod == 0 {
		scanPeriod = 30
	}
	log.Printf("will scan for expired keys every %v seconds\r\n", scanPeriod)
	wg := new(sync.WaitGroup)
	for {
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
						expires := time.Unix(int64(slot.Expires), 0)
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
					block.WriteToFile(dir)
				}(block, s.Dir)
			}
			wg.Wait()
			time.Sleep(time.Duration(scanPeriod) * time.Second)
		}
	}
}
