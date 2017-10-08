package management

import "github.com/packet-guardian/pg-dhcp/internal/server"

type Network int

func (n *Network) GetNameList(_ int, reply *[]string) error {
	*reply = server.GetNetworkList()
	return nil
}
