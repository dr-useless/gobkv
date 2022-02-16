package main

import (
	"flag"
	"log"
	"net/rpc"
)

var cpuProfile = flag.String("cpuprof", "", "write cpu profile to `file`")
var memProfile = flag.String("memprof", "", "write memory profile to `file`")

var configFile = flag.String("c", "", "must be a file path")

func main() {
	flag.Parse()
	log.SetPrefix("gobkv ")

	// pprof
	startCPUProfile()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("failed to load config ", err)
	}

	store := Store{
		AuthSecret: cfg.AuthSecret,
	}
	store.ensureParts(&cfg)

	watchdog := Watchdog{
		Store: &store,
		Cfg:   &cfg,
	}
	watchdog.readFromPartFiles()
	go watchdog.watch()

	rpc.Register(&store)

	listener, err := getListener(&cfg)
	if err != nil {
		log.Fatal("failed to get listener: ", err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("failed to accept conn: ", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
