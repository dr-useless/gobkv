package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/intob/rocketkv/repl"
	"github.com/intob/rocketkv/store"
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
	BufferSize       int    // max Msg length
	Repl             repl.ReplConfig
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

func (c *Config) setDefaults() bool {
	applied := false
	if c.BufferSize < 1 {
		c.BufferSize = 2000000 // 2MB
		applied = true
	}
	if c.ExpiryScanPeriod < 1 {
		c.ExpiryScanPeriod = 10
		applied = true
	}
	if c.Parts.Count < 1 {
		c.Parts.Count = 16
		applied = true
	}
	if c.Parts.Persist && c.Parts.WritePeriod < 1 {
		c.Parts.WritePeriod = 10
		applied = true
	}
	if c.Repl.BufferSize < 1 {
		c.Repl.BufferSize = c.BufferSize
	}
	return applied
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
		defaultsApplied := cfg.setDefaults()
		fmt.Printf("config: %+v\r\n", cfg)
		if defaultsApplied {
			fmt.Println("some defaults applied")
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
