package store

import (
	"net"

	"github.com/packet-guardian/pg-dhcp/models"
)

type Store interface {
	Close() error

	GetLease(ip net.IP) (*models.Lease, error)
	PutLease(l *models.Lease) error
	ForEachLease(foreach func(*models.Lease))

	GetDevice(mac net.HardwareAddr) *models.Device
	PutDevice(d *models.Device)
	DeleteDevice(d *models.Device)
	ForEachDevice(foreach func(*models.Device))
}
