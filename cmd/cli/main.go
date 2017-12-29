package main

import (
	"flag"
	"log"
	"net"
	"time"

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

	t := &timer{}

	if command == "device-test" {
		mac, _ := net.ParseMAC(args[0])

		t.start()
		if err := client.Device().Register(mac); err != nil {
			log.Fatal(err)
		}
		t.print()

		t.start()
		if err := getAndPrintDevice(client, mac); err != nil {
			log.Fatal(err)
		}
		t.print()

		t.start()
		if err := client.Device().Blacklist(mac); err != nil {
			log.Fatal(err)
		}
		t.print()

		t.start()
		if err := getAndPrintDevice(client, mac); err != nil {
			log.Fatal(err)
		}
		t.print()

		t.start()
		if err := client.Device().Unregister(mac); err != nil {
			log.Fatal(err)
		}
		t.print()

		t.start()
		if err := getAndPrintDevice(client, mac); err != nil {
			log.Fatal(err)
		}
		t.print()

		t.start()
		if err := client.Device().RemoveBlacklist(mac); err != nil {
			log.Fatal(err)
		}
		t.print()

		t.start()
		if err := getAndPrintDevice(client, mac); err != nil {
			log.Fatal(err)
		}
		t.print()
	} else if command == "network-test" {
		list, err := client.Network().GetNameList()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%#v\n", list)
	} else if command == "lease-test" {
		leases, err := client.Lease().GetAllFromNetwork("network1")
		if err != nil {
			log.Fatal(err)
		}
		for _, lease := range leases {
			log.Printf("%#v\n", lease)
		}

		lease, err := client.Lease().Get(net.ParseIP("10.0.2.12"))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("\n%#v\n", lease)

		lease, err = client.Lease().Get(net.ParseIP("10.0.2.3"))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("\n%#v\n", lease)
	} else if command == "stat-test" {
		stats, err := client.Server().GetPoolStats()
		if err != nil {
			log.Fatal(err)
		}
		for _, stat := range stats {
			log.Printf("%#v\n", stat)
		}
	}
}

func getAndPrintDevice(c rpcclient.Client, mac net.HardwareAddr) error {
	d, err := c.Device().Get(mac)
	if err != nil {
		return err
	}

	log.Printf("%#v\n", d)
	return nil
}

type timer struct {
	s time.Time
}

func (t *timer) start() {
	t.s = time.Now()
}

func (t *timer) print() {
	log.Printf("%s\n", time.Now().Sub(t.s).String())
}
