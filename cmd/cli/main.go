package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"text/template"
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

	switch command {
	case "leases":
		getLeases(client, args)
	case "networks":
		getNetworkNames(client)
	case "pools":
		getPoolStats(client)
	case "devices":
		devicesCmd(client, args)
	default:
		fmt.Printf("\"%s\" is not a command\n", command)
		os.Exit(1)
	}
}

var multLleaseTemplate = template.Must(template.New("").Parse(`Server Time: {{.Now.Format "2006-01-02 15:04:05 -07:00"}}

Leases in {{.Network}}:
{{range .Leases}}
	IP:         {{.IP.String}}
	MAC:        {{.MAC.String}}
	Start:      {{.Start.Format "2006-01-02 15:04:05 -07:00"}}
	End:        {{.End.Format "2006-01-02 15:04:05 -07:00"}}
	Hostname:   {{.Hostname}}
	Registered: {{.Registered}}
{{end}}
`))

var singleLeaseTemplate = template.Must(template.New("").Parse(`Server Time: {{.Now.Format "2006-01-02 15:04:05 -07:00"}}
{{with .Lease}}
	IP:         {{.IP.String}}
	MAC:        {{.MAC.String}}
	Start:      {{.Start.Format "2006-01-02 15:04:05 -07:00"}}
	End:        {{.End.Format "2006-01-02 15:04:05 -07:00"}}
	Hostname:   {{.Hostname}}
	Registered: {{.Registered}}
{{end}}
`))

func getLeases(client rpcclient.Client, args []string) {
	fs := flag.NewFlagSet("leases", flag.ExitOnError)
	network := fs.String("n", "", "Network")
	address := fs.String("ip", "", "IP Address")
	fs.Parse(args)

	if *network != "" {
		leases, err := client.Lease().GetAllFromNetwork(*network)
		if err != nil {
			log.Fatal(err)
		}
		if leases == nil {
			fmt.Printf("Network %s doesn't exist\n", *network)
			os.Exit(1)
		}

		multLleaseTemplate.Execute(os.Stdout, map[string]interface{}{
			"Now":     time.Now(),
			"Network": network,
			"Leases":  leases,
		})
	} else if *address != "" {
		lease, err := client.Lease().Get(net.ParseIP(*address))
		if err != nil {
			log.Fatal(err)
		}
		if lease == nil {
			fmt.Printf("Lease for %s doesn't exist\n", *address)
			os.Exit(1)
		}

		singleLeaseTemplate.Execute(os.Stdout, map[string]interface{}{
			"Now":   time.Now(),
			"Lease": lease,
		})
	} else {
		fs.PrintDefaults()
		os.Exit(1)
	}
}

func getNetworkNames(client rpcclient.Client) {
	networks, err := client.Network().GetNameList()
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(networks)
	fmt.Println(strings.Join(networks, "\n"))
}

var poolStatsTemplate = template.Must(template.New("").Parse(`Server Time: {{.Now.Format "2006-01-02 15:04:05 -07:00"}}

Pool Statistics:
{{range .Pools}}
	Network:     {{.NetworkName}}
	Subnet:      {{.Subnet}}
	Start:       {{.Start}}
	End:         {{.End}}
	Registered:  {{.Registered}}
	Total:       {{.Total}}
	Active:      {{.Active}}
	Claimed:     {{.Claimed}}
	Abandoned:   {{.Abandoned}}
	Free:        {{.Free}}
{{end}}
`))

func getPoolStats(client rpcclient.Client) {
	stats, err := client.Server().GetPoolStats()
	if err != nil {
		log.Fatal(err)
	}

	poolStatsTemplate.Execute(os.Stdout, map[string]interface{}{
		"Now":   time.Now(),
		"Pools": stats,
	})
}

func devicesCmd(client rpcclient.Client, args []string) {
	if len(args) != 2 {
		fmt.Println("Usage: devices [show|register|unregister|blacklist|unblacklist|delete] MAC")
		os.Exit(1)
	}

	cmd := args[0]
	mac, err := net.ParseMAC(args[1])
	if err != nil {
		fmt.Println("Invalid MAC address")
		os.Exit(1)
	}

	switch cmd {
	case "show":
		devicesCmdShow(client, mac)
	case "register":
		devicesCmdRegister(client, mac)
	case "unregister":
		devicesCmdUnregister(client, mac)
	case "blacklist":
		devicesCmdBlacklist(client, mac)
	case "unblacklist":
		devicesCmdUnblacklist(client, mac)
	case "delete":
		devicesCmdDelete(client, mac)
	default:
		fmt.Println("Usage: devices [show|register|unregister|blacklist|unblacklist|delete] MAC")
		os.Exit(1)
	}
}

var singleDeviceTemplate = template.Must(template.New("").Parse(`{{with .Device}}
	MAC:         {{.MAC.String}}
	Registered:  {{.Registered}}
	Blacklisted: {{.Blacklisted}}
{{end}}
`))

func devicesCmdShow(client rpcclient.Client, mac net.HardwareAddr) {
	device, err := client.Device().Get(mac)
	if err != nil {
		log.Fatal(err)
	}

	singleDeviceTemplate.Execute(os.Stdout, map[string]interface{}{"Device": device})
}

func devicesCmdRegister(client rpcclient.Client, mac net.HardwareAddr) {
	if err := client.Device().Register(mac); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s registered successfully\n", mac.String())
}

func devicesCmdUnregister(client rpcclient.Client, mac net.HardwareAddr) {
	if err := client.Device().Unregister(mac); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s unregistered successfully\n", mac.String())
}

func devicesCmdBlacklist(client rpcclient.Client, mac net.HardwareAddr) {
	if err := client.Device().Blacklist(mac); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s blacklisted successfully\n", mac.String())
}

func devicesCmdUnblacklist(client rpcclient.Client, mac net.HardwareAddr) {
	if err := client.Device().RemoveBlacklist(mac); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s unblacklisted successfully\n", mac.String())
}

func devicesCmdDelete(client rpcclient.Client, mac net.HardwareAddr) {
	if err := client.Device().Delete(mac); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s deleted successfully\n", mac.String())
}
