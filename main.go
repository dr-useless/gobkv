package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/intob/rocketkv/cfg"
	"github.com/intob/rocketkv/store"
	"github.com/intob/rocketkv/util"
	"github.com/spf13/viper"
)

func main() {
	cfg.InitConfig()

	st := store.NewStore()

	listener, err := util.GetListener(
		viper.GetString(cfg.NETWORK),
		viper.GetString(cfg.ADDRESS),
		viper.GetString(cfg.TLS_CERT),
		viper.GetString(cfg.TLS_KEY))
	if err != nil {
		panic(err)
	}

	dir := viper.GetString(cfg.DIR)
	go waitForSigInt(listener, st, dir)

	// repl
	if viper.GetBool(cfg.REPL_ENABLED) {
		// repl.NewReplService(st)
		log.Fatal("replication not yet (re-)implemented")
	}

	auth := viper.GetString(cfg.AUTH)
	bufSize := viper.GetInt(cfg.BUFFER_SIZE)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go st.ServeConn(conn, auth, bufSize)
	}
}

func waitForSigInt(listener net.Listener, st *store.Store, dir string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		fmt.Println("will exit cleanly")
		listener.Close()
		st.WriteAllBlocks(dir)
		os.Exit(0)
	}
}
