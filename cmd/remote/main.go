package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

var (
	listenAddr string
)

func init() {
	flag.StringVar(&listenAddr, "a", "127.0.0.1:5925", "Listen address and port")
}

func main() {
	flag.Parse()

	conn, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer conn.Close()

	for {
		client, err := conn.Accept()
		if err != nil {
			log.Fatal(err.Error())
		}

		log.Println("Accepting client " + client.RemoteAddr().String())
		go accept(client)
	}
}

func accept(conn net.Conn) {
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err.Error())
			return
		}

		thisBuf := buf[:n]
		fmt.Printf("%v\n", thisBuf)
		conn.Write([]byte{1, 'R', 'R'})
	}
}
