package rpcclient

import (
	"net"

	"github.com/packet-guardian/pg-dhcp/models"
)

type LeaseRPCRequest struct {
	client *RPCClient
}

func (l *LeaseRPCRequest) GetAllFromNetwork(name string) ([]*models.Lease, error) {
	var reply []*models.Lease
	if err := l.client.call("Lease.GetAllFromNetwork", name, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (l *LeaseRPCRequest) Get(ip net.IP) (*models.Lease, error) {
	reply := new(models.Lease)
	if err := l.client.call("Lease.Get", ip, reply); err != nil || reply.IP == nil {
		return nil, err
	}
	return reply, nil
}
