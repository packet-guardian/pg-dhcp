package management

import (
	"net"

	"github.com/packet-guardian/pg-dhcp/internal/server"
	"github.com/packet-guardian/pg-dhcp/models"
	"github.com/packet-guardian/pg-dhcp/store"
)

type Lease struct {
	store store.Store
}

func (l *Lease) GetAllFromNetwork(name string, reply *[]*models.Lease) error {
	*reply = server.GetLeasesInNetwork(name)
	return nil
}

func (l *Lease) Get(ip net.IP, reply *models.Lease) error {
	lease, _ := l.store.GetLease(ip)
	if lease == nil {
		return nil
	}

	*reply = *lease
	return nil
}
