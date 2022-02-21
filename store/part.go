package store

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"hash/fnv"
	"log"
	"os"
	"path"
	"sync"
)

const listFileName = "parts.gob"
const hashLen = 4

type Part struct {
	Id        []byte
	Mutex     *sync.RWMutex
	Data      map[string]*Slot
	MustWrite bool
}

type Slot struct {
	Value    []byte
	Expires  uint64
	Modified uint64
}

type PartConfig struct {
	Count       int
	Persist     bool
	WritePeriod int // seconds
}

func (part *Part) WriteToFile(partName, dir string) {
	if !part.MustWrite {
		return
	}
	fullPath := path.Join(dir, partName+".gob")
	file, err := os.Create(fullPath)
	if err != nil {
		log.Fatalf("failed to create part file: %s\r\n", err)
	}
	part.Mutex.RLock()
	gob.NewEncoder(file).Encode(&part.Data)
	part.MustWrite = false
	part.Mutex.RUnlock()
	file.Close()
}

// ensures that part files exist
func (s *Store) EnsureParts(cfg *PartConfig) {
	if !cfg.Persist {
		return
	}
	s.Parts = make(map[string]*Part)
	listPath := path.Join(s.Dir, listFileName)
	listFile, err := os.Open(listPath)
	if err != nil {
		log.Println("no part list found, will create...")
		// make new list
		for i := 0; i < cfg.Count; i++ {
			partId := make([]byte, hashLen)
			rand.Read(partId)
			partName := getPartName(partId)
			s.Parts[partName] = &Part{
				Id:    partId,
				Mutex: new(sync.RWMutex),
				Data:  make(map[string]*Slot),
			}
		}
		newListFile, err := os.Create(listPath)
		if err != nil {
			log.Fatalf("failed to create part list, check directory exists: %s", s.Dir)
		}
		partNameList := s.getPartNameList()
		gob.NewEncoder(newListFile).Encode(partNameList)
	} else {
		// decode list
		nameList := make([]string, 0)
		err := gob.NewDecoder(listFile).Decode(&nameList)
		if err != nil {
			log.Fatalf("failed to decode part list: %s", err)
		}
		for _, name := range nameList {
			id, _ := getPartId(name)
			s.Parts[name] = &Part{
				Id:    id,
				Mutex: new(sync.RWMutex),
				Data:  make(map[string]*Slot),
			}
		}
		log.Printf("initialised %v parts from list\r\n", len(s.Parts))
	}
}

func (p *Part) ReadFromFile(wg *sync.WaitGroup, dir string) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	defer wg.Done()
	name := getPartName(p.Id)
	fullPath := path.Join(dir, name+".gob")
	file, err := os.Open(fullPath)
	if err != nil {
		log.Printf("failed to open partition %s\r\n", name)
		return
	}
	err = gob.NewDecoder(file).Decode(&p.Data)
	if err != nil {
		log.Printf("failed to decode data in partition %s\r\n", name)
		return
	}
	log.Printf("read from partition %s", name)
}

func (s *Store) getClosestPart(key string) *Part {
	h := fnv.New32a()
	h.Write([]byte(key))
	keyHash := h.Sum(nil)
	var clPart *Part
	var clD []byte
	for _, part := range s.Parts {
		d := xorBytes(part.Id, keyHash)
		if clD == nil || bytes.Compare(d, clD) < 0 {
			clPart = part
			clD = d
		}
	}
	return clPart
}

func (s *Store) getPartNameList() []string {
	list := make([]string, 0)
	for name := range s.Parts {
		list = append(list, name)
	}
	return list
}

func getPartName(partId []byte) string {
	return base64.RawURLEncoding.EncodeToString(partId)
}

func getPartId(partName string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(partName)
}
