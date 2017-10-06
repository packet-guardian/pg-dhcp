package management

import "github.com/packet-guardian/pg-dhcp/internal/server"

type Network int

func (n *Network) GetNameList() []string {
	return server.GetNetworkList()
}
