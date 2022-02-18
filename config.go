package main

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	Network          string // tcp, unix etc...
	Address          string // 0.0.0.0:8100
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
	if c.Network == "" || c.Address == "" {
		return errors.New("network & address must not be blank")
	}
	if c.Persist {
		if c.PartCount < 1 {
			return errors.New("part count must be greater than 0")
		}
	}
	return nil
}

func loadConfig() (Config, error) {
	cfg := Config{}

	if *configFile == "" {
		return cfg, errors.New("no config file provided")
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
