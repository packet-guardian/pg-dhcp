package main

import (
	"flag"
	"log"
	"net"

	"github.com/packet-guardian/pg-dhcp/rpcclient"
)

var (
	serverAddress string
)

func init() {
	flag.StringVar(&serverAddress, "h", "localhost:8677", "DHCP managment host address")
}

func main() {
	flag.Parse()

	client, err := rpcclient.Connect("tcp", serverAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	command := flag.Arg(0)
	var args []string
	if len(flag.Args()) > 0 {
		args = flag.Args()[1:]
	}

	if command == "device-test" {
		mac, _ := net.ParseMAC(args[0])
		if err := client.Device().Register(mac); err != nil {
			log.Fatal(err)
		}

		d, err := client.Device().Get(mac)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%#v\n", d)
	}
}
