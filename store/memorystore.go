package store

import (
	"net"
	"sync"

	"github.com/packet-guardian/pg-dhcp/models"
)

type MemoryStore struct {
	m       sync.RWMutex
	leases  map[string]*models.Lease
	devices map[string]*models.Device
}

func NewMemoryStore() (*MemoryStore, error) {
	s := &MemoryStore{
		leases:  make(map[string]*models.Lease),
		devices: make(map[string]*models.Device),
	}
	return s, nil
}

func (s *MemoryStore) Close() error {
	return nil
}

func (s *MemoryStore) GetLease(ip net.IP) (*models.Lease, error) {
	var l *models.Lease
	s.m.RLock()
	l = s.leases[ip.String()]
	s.m.RUnlock()
	return l, nil
}

func (s *MemoryStore) PutLease(l *models.Lease) error {
	s.m.Lock()
	s.leases[l.IP.String()] = l
	s.m.Unlock()
	return nil
}

func (s *MemoryStore) ForEachLease(foreach func(*models.Lease)) error {
	s.m.RLock()
	for _, v := range s.leases {
		foreach(v)
	}
	s.m.RUnlock()
	return nil
}

func (s *MemoryStore) GetDevice(mac net.HardwareAddr) (*models.Device, error) {
	var d *models.Device
	s.m.RLock()
	d = s.devices[mac.String()]
	s.m.RUnlock()

	if d == nil {
		d = &models.Device{
			MAC:         mac,
			Registered:  false,
			Blacklisted: false,
		}
	}
	return d, nil
}

func (s *MemoryStore) PutDevice(d *models.Device) error {
	s.m.Lock()
	s.devices[d.MAC.String()] = d
	s.m.Unlock()
	return nil
}

func (s *MemoryStore) DeleteDevice(d *models.Device) error {
	s.m.Lock()
	delete(s.devices, d.MAC.String())
	s.m.Unlock()
	return nil
}

func (s *MemoryStore) ForEachDevice(foreach func(*models.Device)) error {
	s.m.RLock()
	for _, v := range s.devices {
		foreach(v)
	}
	s.m.RUnlock()
	return nil
}
