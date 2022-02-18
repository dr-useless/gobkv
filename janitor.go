package main

import (
	"log"
	"sync"
	"time"
)

// Delete expired keys
// TODO: consider sleeping before wait,
// this would ensure a (more) constant period between scans
func (s *Store) scanForExpiredKeys(cfg *Config) {
	if cfg.ExpiryScanPeriod == 0 {
		cfg.ExpiryScanPeriod = 30
	}
	log.Printf("will scan for expired keys every %v seconds\r\n", cfg.ExpiryScanPeriod)
	wg := new(sync.WaitGroup)
	for {
		for partName, part := range s.Parts {
			wg.Add(1)
			go func(part *Part, partName string, partDir string) {
				defer wg.Done()
				part.Mux.RLock()
				for k, v := range part.Data {
					if v.Expires == 0 {
						continue
					}
					expires := time.Unix(int64(v.Expires), 0)
					if time.Now().After(expires) {
						part.Mux.RUnlock()
						part.Mux.Lock()
						delete(part.Data, k)
						part.MustWrite = true
						part.Mux.Unlock()
						part.Mux.RLock()
					}
				}
				part.Mux.RUnlock()
				part.writeToFile(partName, cfg.PartDir)
			}(part, partName, cfg.PartDir)
		}
		wg.Wait()
		time.Sleep(time.Duration(cfg.ExpiryScanPeriod) * time.Second)
	}
}
