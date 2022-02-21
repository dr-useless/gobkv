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
		for partName, part := range s.Parts {
			wg.Add(1)
			go func(part *Part, partName string, partDir string) {
				defer wg.Done()
				part.Mutex.RLock()
				for k, slot := range part.Slots {
					if slot.Expires == 0 {
						continue
					}
					expires := time.Unix(int64(slot.Expires), 0)
					if time.Now().After(expires) {
						part.Mutex.RUnlock()
						part.Mutex.Lock()
						delete(part.Slots, k)
						part.MustWrite = true
						part.Mutex.Unlock()
						part.Mutex.RLock()
					}
				}
				part.Mutex.RUnlock()
				part.WriteToFile(partName, s.Dir)
			}(part, partName, s.Dir)
		}
		wg.Wait()
		time.Sleep(time.Duration(scanPeriod) * time.Second)
	}
}
