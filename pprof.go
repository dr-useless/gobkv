package main

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

var cpuProfFile *os.File

func startCPUProfile() {
	if *cpuProfile == "" {
		return
	}
	var err error
	cpuProfFile, err = os.Create(*cpuProfile)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(cpuProfFile); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
}

func stopCPUProfile() {
	if *cpuProfile == "" {
		return
	}
	pprof.StopCPUProfile()
	if err := cpuProfFile.Close(); err != nil {
		log.Fatal("error closing cpuProfFile: ", err)
	}
}

func makeMemProfile() {
	if *memProfile == "" {
		return
	}
	f, err := os.Create(*memProfile)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	runtime.GC()    // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
}
