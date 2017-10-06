package management

import (
	"net"

	"github.com/packet-guardian/pg-dhcp/internal/server"
	"github.com/packet-guardian/pg-dhcp/store"
)

type Lease struct {
	store *store.Store
}

func (l *Lease) GetAllFromNetwork(name string) []*store.Lease {
	return server.GetLeasesInNetwork(name)
}

func (l *Lease) Get(ip net.IP) *store.Lease {
	lease, _ := l.store.GetLease(ip)
	return lease
}
