// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"net"
	"time"
)

type testLeaseStore struct{}

func (s *testLeaseStore) GetAllLeases() ([]*Lease, error)                          { return nil, nil }
func (s *testLeaseStore) GetLeaseByIP(ip net.IP) (*Lease, error)                   { return nil, nil }
func (s *testLeaseStore) GetRecentLeaseByMAC(mac net.HardwareAddr) (*Lease, error) { return nil, nil }
func (s *testLeaseStore) GetAllLeasesByMAC(mac net.HardwareAddr) ([]*Lease, error) { return nil, nil }
func (s *testLeaseStore) CreateLease(l *Lease) error                               { return nil }
func (s *testLeaseStore) UpdateLease(l *Lease) error                               { return nil }
func (s *testLeaseStore) DeleteLease(l *Lease) error                               { return nil }
func (s *testLeaseStore) SearchLeases(where string, vals ...interface{}) ([]*Lease, error) {
	return nil, nil
}

type testDeviceStore struct {
	macs map[string]Device
	next Device
}

func (d *testDeviceStore) GetDeviceByMAC(mac net.HardwareAddr) (Device, error) {
	if d.next != nil {
		return d.next, nil
	}
	return &testDevice{store: d, mac: mac}, nil
}
func (d *testDeviceStore) setNextDevice(t Device) { d.next = t }

type testDevice struct {
	store       *testDeviceStore
	lastTime    time.Time
	id          int
	mac         net.HardwareAddr
	username    string
	blacklisted bool
	expired     bool
	registered  bool
}

func (d *testDevice) SetLastSeen(t time.Time)  { d.lastTime = t }
func (d *testDevice) GetID() int               { return d.id }
func (d *testDevice) GetMAC() net.HardwareAddr { return d.mac }
func (d *testDevice) GetUsername() string      { return d.username }
func (d *testDevice) IsBlacklisted() bool      { return d.blacklisted }
func (d *testDevice) IsExpired() bool          { return d.expired }
func (d *testDevice) IsRegistered() bool       { return d.registered }
func (d *testDevice) Save() error              { return nil }
