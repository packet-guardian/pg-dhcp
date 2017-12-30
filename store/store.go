package store

import (
	"net"

	"github.com/packet-guardian/pg-dhcp/models"
)

type Store interface {
	Close() error

	GetLease(ip net.IP) (*models.Lease, error)
	PutLease(l *models.Lease) error
	ForEachLease(foreach func(*models.Lease)) error

	GetDevice(mac net.HardwareAddr) (*models.Device, error)
	PutDevice(d *models.Device) error
	DeleteDevice(d *models.Device) error
	ForEachDevice(foreach func(*models.Device)) error
}
