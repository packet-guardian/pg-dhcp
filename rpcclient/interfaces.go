package rpcclient

import (
	"net"

	"github.com/packet-guardian/pg-dhcp/models"
	"github.com/packet-guardian/pg-dhcp/stats"
)

type Client interface {
	Close() error
	Device() DeviceRequest
	Lease() LeaseRequest
	Network() NetworkRequest
	Server() ServerRequest
}

type DeviceRequest interface {
	Get(mac net.HardwareAddr) (*models.Device, error)
	Register(mac net.HardwareAddr) error
	Unregister(mac net.HardwareAddr) error
	Blacklist(mac net.HardwareAddr) error
	RemoveBlacklist(mac net.HardwareAddr) error
	Delete(mac net.HardwareAddr) error
}

type LeaseRequest interface {
	GetAllFromNetwork(name string) ([]*models.Lease, error)
	Get(ip net.IP) (*models.Lease, error)
}

type NetworkRequest interface {
	GetNameList() ([]string, error)
}

type ServerRequest interface {
	GetPoolStats() ([]*stats.PoolStat, error)
}
