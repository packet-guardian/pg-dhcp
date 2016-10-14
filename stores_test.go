// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"net"
	"time"
)

var testConfig = `
global
	option domain-name example.com

	server-identifier 10.0.0.1

	registered
		default-lease-time 86400
		max-lease-time 86400
		option domain-name-server 10.1.0.1 10.1.0.2
	end

	unregistered
		default-lease-time 360
		max-lease-time 360
		option domain-name-server 10.0.0.1
	end
end

network network1
	unregistered
		subnet 10.0.1.0/24
			range 10.0.1.10 10.0.1.200
			option router 10.0.1.1
		end
	end
	registered
		subnet 10.0.2.0/24
			range 10.0.2.10 10.0.2.200
			option router 10.0.2.1
		end
	end
end

network network2
	unregistered
		subnet 10.0.4.0/22
			range 10.0.4.1 10.0.7.254
			option router 10.0.4.1
		end
	end
	registered
		subnet 10.0.3.0/24
			range 10.0.3.10 10.0.3.200
			option router 10.0.3.1
		end
	end
end

network network3
	unregistered
		subnet 10.0.8.0/24
			pool
				range 10.0.8.10 10.0.8.100
			end
			pool
				range 10.0.8.120 10.0.8.250
			end
		end
	end
end

network network4
	unregistered
		subnet 10.0.9.0/24
			range 10.0.9.10 10.0.9.100
			range 10.0.9.120 10.0.9.250
		end
	end
end
`

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
