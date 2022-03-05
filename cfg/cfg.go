package cfg

import (
	"flag"
	"fmt"

	"github.com/spf13/viper"
)

const NETWORK = "network"   // listen network
const ADDRESS = "address"   // listen address
const TLS_CERT = "tls.cert" // TLS cert file
const TLS_KEY = "tls.key"   // TLS key file
const AUTH = "auth"         // auth secret

const SEGMENTS = "segments"      // number of parts & blocks
const BUFFER_SIZE = "buffersize" // maximum length of a single message (including value)
const SCAN_PERIOD = "scanperiod" // seconds between scanning for expired keys

const PERSIST = "persist" // bool
// if persist = true:
const WRITE_PERIOD = "writeperiod" // seconds between writing changed blocks to file
const DIR = "dir"                  // directory for blocks

const REPL_ENABLED = "repl.enabled"
const REPL_NETWORK = "repl.network"
const REPL_ADDRESS = "repl.address"
const REPL_TLS_CERT = "repl.tls.cert"
const REPL_TLS_KEY = "repl.tls.key"
const REPL_NAME = "repl.name"
const REPL_PERIOD = "repl.period"

var configFile = flag.String("c", "", "must be a file path")

// InitConfig loads a config file using Viper
func InitConfig() {
	initDefaults()

	viper.SetConfigName("config")

	flag.Parse()
	if *configFile != "" {
		viper.SetConfigFile(*configFile)
	} else {
		viper.AddConfigPath("/etc/rocketkv/")
		viper.AddConfigPath(".")
	}

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

// Start with sensible defaults
func initDefaults() {
	viper.SetDefault(NETWORK, "tcp")
	viper.SetDefault(ADDRESS, ":8100")

	viper.SetDefault(BUFFER_SIZE, 2000000) // 2MB
	viper.SetDefault(SEGMENTS, "16")       // 256 blocks
	viper.SetDefault(SCAN_PERIOD, 10)

	viper.SetDefault(WRITE_PERIOD, 10)
	viper.SetDefault(DIR, ".")
}
