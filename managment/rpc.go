package management

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/lfkeitel/verbose/v5"
	"github.com/packet-guardian/pg-dhcp/internal/config"
	"github.com/packet-guardian/pg-dhcp/store"
)

var logger *verbose.Logger

func SetLogger(l *verbose.Logger) { logger = l }

// StartRPCServer starts a managment RPC server connection
func StartRPCServer(c *config.ManagementConfig, db store.Store) error {
	registerStructs(db)

	addr := fmt.Sprintf("%s:%d", c.Address, c.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	logger.WithField("address", addr).Info("Management server listening")

	return serve(l, c.AllowedIPs)
}

func registerStructs(db store.Store) {
	rpc.Register(new(Network))
	rpc.Register(new(Server))
	rpc.Register(&Lease{store: db})
	rpc.Register(&Device{store: db})
}

func serve(l net.Listener, allowedIPs []string) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		if allowedIP(conn.RemoteAddr(), allowedIPs) {
			go rpc.DefaultServer.ServeConn(conn)
		} else {
			conn.Close()
			logger.WithField("address", conn.RemoteAddr().String()).Info("Blocked management request")
		}
	}
}

func allowedIP(a net.Addr, ips []string) bool {
	if len(ips) == 0 {
		return true
	}

	host, _, err := net.SplitHostPort(a.String())
	if err != nil {
		return false
	}

	for _, ip := range ips {
		if host == ip {
			return true
		}
	}
	return false
}
