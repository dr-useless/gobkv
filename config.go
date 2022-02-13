package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

type Config struct {
	Port       int
	CertFile   string
	KeyFile    string
	AuthSecret string
}

func loadConfig() (Config, error) {
	configFile := ""
	flag.StringVar(&configFile, "c", configFile, "must be a file path")
	flag.Parse()

	cfg := Config{
		Port: 8100,
	}

	if configFile == "" {
		log.Println("no config file defined, running with defaults")
		return cfg, nil
	} else {
		err := read(configFile, &cfg)
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
