package rpcclient

import (
	"net"

	"github.com/packet-guardian/pg-dhcp/store"
)

type LeaseRequest struct {
	client *Client
}

func (l *LeaseRequest) GetAllFromNetwork(name string) ([]*store.Lease, error) {
	reply := make([]*store.Lease, 0)
	if err := l.client.c.Call("Lease.GetAllFromNetwork", nil, reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (l *LeaseRequest) Get(ip net.IP) (*store.Lease, error) {
	reply := new(store.Lease)
	if err := l.client.c.Call("Lease.Get", nil, reply); err != nil {
		return nil, err
	}
	return reply, nil
}
