package main

import (
	"flag"
	"fmt"
	"log"
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
		if len(args) != 1 {
			fmt.Println("leases command expected a network name")
			os.Exit(1)
		}
		getLeases(client, args[0])
	case "networks":
		getNetworkNames(client)
	default:
		fmt.Printf("\"%s\" is not a command\n", command)
		os.Exit(1)
	}
}

var leasesTemplate = template.Must(template.New("").Parse(`Server Time: {{.Now.Format "2006-01-02 15:04:05 07:00"}}

Leases in {{.Network}}:
{{range .Leases}}
	IP:         {{.IP.String}}
	MAC:        {{.MAC.String}}
	Start:      {{.Start.Format "2006-01-02 15:04:05 07:00"}}
	End:        {{.End.Format "2006-01-02 15:04:05 07:00"}}
	Hostname:   {{.Hostname}}
	Registered: {{.Registered}}
{{end}}
`))

func getLeases(client rpcclient.Client, network string) {
	leases, err := client.Lease().GetAllFromNetwork(network)
	if err != nil {
		log.Fatal(err)
	}

	leasesTemplate.Execute(os.Stdout, map[string]interface{}{
		"Now":     time.Now(),
		"Network": network,
		"Leases":  leases,
	})
}

func getNetworkNames(client rpcclient.Client) {
	networks, err := client.Network().GetNameList()
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(networks)
	fmt.Println(strings.Join(networks, "\n"))
}
