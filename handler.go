package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		data, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			log.Println("end ", err)
			return
		}

		log.Println("got ", data)

		_, err = c.Write([]byte(data))
		if err != nil {
			log.Println("write error: ", err)
			return
		}
	}
}
