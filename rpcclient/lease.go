package rpcclient

import (
	"net"

	"github.com/packet-guardian/pg-dhcp/store"
)

type LeaseRequest struct {
	client *Client
}

func (l *LeaseRequest) GetAllFromNetwork(name string) ([]*store.Lease, error) {
	var reply []*store.Lease
	if err := l.client.c.Call("Lease.GetAllFromNetwork", name, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (l *LeaseRequest) Get(ip net.IP) (*store.Lease, error) {
	reply := new(store.Lease)
	if err := l.client.c.Call("Lease.Get", ip, reply); err != nil {
		return nil, err
	}
	return reply, nil
}
