package management

import (
	"net"
	"net/rpc"

	"github.com/packet-guardian/pg-dhcp/store"
)

// StartRPCServer starts a managment RPC server connection
func StartRPCServer(l net.Listener, db store.Store) {
	rpc.Register(new(Network))
	rpc.Register(new(Server))
	rpc.Register(&Lease{store: db})
	rpc.Register(&Device{store: db})
	rpc.Accept(l)
}
