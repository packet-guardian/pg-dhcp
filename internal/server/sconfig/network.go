// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sconfig

import (
	"bytes"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/packet-guardian/pg-dhcp/models"
)

type Network struct {
	sync.Mutex
	global               *Global
	Name                 string
	settings             *Settings
	registeredSettings   *Settings
	regOptionsCached     bool
	unregisteredSettings *Settings
	unregOptionsCached   bool
	Subnets              []*Subnet
	IgnoreRegistration   bool
	EnforceBlocklist     bool
}

func newNetwork(name string) *Network {
	return &Network{
		Name:                 strings.ToLower(name),
		settings:             newSettingsBlock(),
		registeredSettings:   newSettingsBlock(),
		unregisteredSettings: newSettingsBlock(),
	}
}

// GetLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If the network does not have an explicitly set duration for either,
// it will get the duration from Global.
func (n *Network) getLeaseTime(req time.Duration, registered bool) time.Duration {
	if req == 0 {
		return n.getDefaultLeaseTime(registered)
	}
	return n.getMaxLeaseTime(req, registered)
}

func (n *Network) getDefaultLeaseTime(registered bool) time.Duration {
	if registered {
		if n.registeredSettings.DefaultLeaseTime > 0 {
			return n.registeredSettings.DefaultLeaseTime
		}
		if n.settings.DefaultLeaseTime > 0 {
			return n.settings.DefaultLeaseTime
		}
		// Save to return early next time
		n.registeredSettings.DefaultLeaseTime = n.global.getLeaseTime(0, registered)
		return n.registeredSettings.DefaultLeaseTime
	}

	if n.unregisteredSettings.DefaultLeaseTime > 0 {
		return n.unregisteredSettings.DefaultLeaseTime
	}
	if n.settings.DefaultLeaseTime > 0 {
		return n.settings.DefaultLeaseTime
	}
	// Save to return early next time
	n.unregisteredSettings.DefaultLeaseTime = n.global.getLeaseTime(0, registered)
	return n.unregisteredSettings.DefaultLeaseTime
}

func (n *Network) getMaxLeaseTime(req time.Duration, registered bool) time.Duration {
	// Registered devices
	if registered {
		if n.registeredSettings.MaxLeaseTime > 0 {
			if req <= n.registeredSettings.MaxLeaseTime {
				return req
			}
			return n.registeredSettings.MaxLeaseTime
		}
		if n.settings.MaxLeaseTime > 0 {
			if req <= n.settings.MaxLeaseTime {
				return req
			}
			return n.settings.MaxLeaseTime
		}
		return n.global.getLeaseTime(req, registered)
	}

	// Unregistered devices
	if n.unregisteredSettings.MaxLeaseTime > 0 {
		if req <= n.unregisteredSettings.MaxLeaseTime {
			return req
		}
		return n.unregisteredSettings.MaxLeaseTime
	}
	if n.settings.MaxLeaseTime > 0 {
		if req <= n.settings.MaxLeaseTime {
			return req
		}
		return n.settings.MaxLeaseTime
	}
	return n.global.getLeaseTime(req, registered)
}

func (n *Network) getSettings(registered bool) *Settings {
	if registered && n.regOptionsCached {
		return n.registeredSettings
	} else if !registered && n.unregOptionsCached {
		return n.unregisteredSettings
	}

	gSet := n.global.getSettings(registered)
	if registered {
		mergeSettings(n.registeredSettings, gSet)
		n.regOptionsCached = true
		return n.registeredSettings
	}

	mergeSettings(n.unregisteredSettings, gSet)
	n.unregOptionsCached = true
	return n.unregisteredSettings
}

func (n *Network) includes(ip net.IP) bool {
	for _, s := range n.Subnets {
		if s.Includes(ip) {
			return true
		}
	}
	return false
}

func (n *Network) GetPoolOfIP(ip net.IP) *Pool {
	for _, s := range n.Subnets {
		for _, p := range s.Pools {
			if p.Includes(ip) {
				return p
			}
		}
	}
	return nil
}

func (n *Network) GetFreeLease(registered bool) (*models.Lease, *Pool) {
	for _, s := range n.Subnets {
		if s.AllowUnknown == registered {
			continue
		}
		for _, p := range s.Pools {
			if l := p.GetFreeLease(); l != nil {
				return l, p
			}
		}
	}
	return nil, nil
}

func (n *Network) GetFreeLeaseDesperate(registered bool) (*models.Lease, *Pool) {
	for _, s := range n.Subnets {
		if s.AllowUnknown == registered {
			continue
		}
		for _, p := range s.Pools {
			if l := p.getFreeLeaseDesperate(); l != nil {
				return l, p
			}
		}
	}
	return nil, nil
}

func (n *Network) GetLeaseByMAC(mac net.HardwareAddr, registered bool) (*models.Lease, *Pool) {
	for _, s := range n.Subnets {
		if s.AllowUnknown == registered {
			continue
		}
		for _, p := range s.Pools {
			for _, l := range p.Leases {
				if bytes.Equal(l.MAC, mac) {
					return l, p
				}
			}
		}
	}
	return nil, nil
}

func (n *Network) GetLeaseByIP(ip net.IP, registered bool) (*models.Lease, *Pool) {
	for _, s := range n.Subnets {
		if s.AllowUnknown == registered {
			continue
		}
		for _, p := range s.Pools {
			if l, ok := p.Leases[ip.String()]; ok {
				return l, p
			}
		}
	}
	return nil, nil
}

func (n *Network) GetAllLeases() []*models.Lease {
	leases := make([]*models.Lease, 0, 20)
	for _, s := range n.Subnets {
		for _, p := range s.Pools {
			for _, l := range p.Leases {
				leases = append(leases, l)
			}
		}
	}
	return leases
}
