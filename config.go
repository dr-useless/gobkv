package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

type Config struct {
	Port             int
	CertFile         string
	KeyFile          string
	AuthSecret       string
	Persist          bool   // write data to file system
	PartCount        int    // number of partitions
	PartDir          string // directory for partition storage, default is ${pwd}/parts
	PartWritePeriod  int    // seconds
	ExpiryScanPeriod int    // seconds
}

func (c *Config) validate() error {
	if c.Persist {
		if c.PartCount < 1 {
			return errors.New("PartCount must be greater than 0")
		}
	}
	return nil
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
		if err != nil {
			return cfg, err
		}
		validityError := cfg.validate()
		return cfg, validityError
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
