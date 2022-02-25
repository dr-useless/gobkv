package main

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/intob/gobkv/store"
)

const ErrNetAddressBlank = "network & address must not be blank"
const ErrZeroPartCount = "part count must be greater than 0"
const ErrMissingConfig = "no config file provided"

type Config struct {
	Network          string // tcp, unix etc...
	Address          string // 0.0.0.0:8100
	CertFile         string
	KeyFile          string
	AuthSecret       string
	Parts            store.PartConfig
	Dir              string // storage dir for blocks
	ExpiryScanPeriod int    // seconds
}

func (c *Config) validate() error {
	if c.Network == "" || c.Address == "" {
		return errors.New(ErrNetAddressBlank)
	}
	if c.Parts.Persist && c.Parts.Count < 1 {
		return errors.New(ErrZeroPartCount)
	}
	return nil
}

func loadConfig() (Config, error) {
	cfg := Config{}

	if *configFile == "" {
		return cfg, errors.New(ErrMissingConfig)
	} else {
		err := read(*configFile, &cfg)
		if err != nil {
			return cfg, err
		}
		return cfg, cfg.validate()
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
