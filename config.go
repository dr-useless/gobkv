package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Port             int
	CertFile         string
	KeyFile          string
	AuthSecret       string
	Persist          bool   // write data to file system
	ShardCount       int    // number of shards used for persistence
	ShardDir         string // directory for shards, default is ${pwd}/shards
	ShardWritePeriod int    // seconds
}

func loadConfig() (Config, error) {
	cfg := Config{
		Port: 8100,
	}

	if *configFile == "" {
		log.Println("no config file defined, running with defaults")
		return cfg, nil
	} else {
		err := read(*configFile, &cfg)
		return cfg, err
	}
}

func read(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}
