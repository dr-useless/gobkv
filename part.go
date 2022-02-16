package main

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
	Mux       *sync.RWMutex
	Data      PartData
	MustWrite bool
}

type PartData map[string][]byte

func (part *Part) writeToFile(partName string, cfg *Config) {
	if !part.MustWrite {
		return
	}
	part.Mux.RLock()
	defer part.Mux.RUnlock()
	fullPath := path.Join(cfg.PartDir, partName+".gob")
	file, err := os.Create(fullPath)
	if err != nil {
		log.Fatalf("failed to create part file: %s\r\n", err)
	}
	defer file.Close()
	gob.NewEncoder(file).Encode(&part.Data)
	part.MustWrite = false
}

// ensures that part files exist
func (s *Store) ensureParts(cfg *Config) {
	if !cfg.Persist {
		return
	}
	s.Parts = make(map[string]*Part)
	listPath := path.Join(cfg.PartDir, listFileName)
	listFile, err := os.Open(listPath)
	if err != nil {
		log.Println("no part list found, will create...")
		// make new list
		for i := 0; i < cfg.PartCount; i++ {
			partId := make([]byte, hashLen)
			rand.Read(partId)
			partName := getPartName(partId)
			s.Parts[partName] = &Part{
				Id:   partId,
				Mux:  new(sync.RWMutex),
				Data: make(map[string][]byte),
			}
		}
		newListFile, err := os.Create(listPath)
		if err != nil {
			log.Fatalf("failed to create part list, check directory exists: %s", cfg.PartDir)
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
				Id:   id,
				Mux:  new(sync.RWMutex),
				Data: make(map[string][]byte),
			}
		}
		log.Printf("initialised %v parts from list\r\n", len(s.Parts))
	}
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
