// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import (
	"bytes"
	"net"
	"strings"
	"time"

	"github.com/packet-guardian/pg-dhcp/store"
)

type Network struct {
	global               *Global
	name                 string
	settings             *settings
	registeredSettings   *settings
	regOptionsCached     bool
	unregisteredSettings *settings
	unregOptionsCached   bool
	subnets              []*subnet
}

func newNetwork(name string) *Network {
	return &Network{
		name:                 strings.ToLower(name),
		settings:             newSettingsBlock(),
		registeredSettings:   newSettingsBlock(),
		unregisteredSettings: newSettingsBlock(),
	}
}

func (n *Network) GetName() string {
	return n.name
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
		if n.registeredSettings.defaultLeaseTime > 0 {
			return n.registeredSettings.defaultLeaseTime
		}
		if n.settings.defaultLeaseTime > 0 {
			return n.settings.defaultLeaseTime
		}
		// Save to return early next time
		n.registeredSettings.defaultLeaseTime = n.global.getLeaseTime(0, registered)
		return n.registeredSettings.defaultLeaseTime
	}

	if n.unregisteredSettings.defaultLeaseTime > 0 {
		return n.unregisteredSettings.defaultLeaseTime
	}
	if n.settings.defaultLeaseTime > 0 {
		return n.settings.defaultLeaseTime
	}
	// Save to return early next time
	n.unregisteredSettings.defaultLeaseTime = n.global.getLeaseTime(0, registered)
	return n.unregisteredSettings.defaultLeaseTime
}

func (n *Network) getMaxLeaseTime(req time.Duration, registered bool) time.Duration {
	// Registered devices
	if registered {
		if n.registeredSettings.maxLeaseTime > 0 {
			if req <= n.registeredSettings.maxLeaseTime {
				return req
			}
			return n.registeredSettings.maxLeaseTime
		}
		if n.settings.maxLeaseTime > 0 {
			if req <= n.settings.maxLeaseTime {
				return req
			}
			return n.settings.maxLeaseTime
		}
		return n.global.getLeaseTime(req, registered)
	}

	// Unregistered devices
	if n.unregisteredSettings.maxLeaseTime > 0 {
		if req <= n.unregisteredSettings.maxLeaseTime {
			return req
		}
		return n.unregisteredSettings.maxLeaseTime
	}
	if n.settings.maxLeaseTime > 0 {
		if req <= n.settings.maxLeaseTime {
			return req
		}
		return n.settings.maxLeaseTime
	}
	return n.global.getLeaseTime(req, registered)
}

func (n *Network) getSettings(registered bool) *settings {
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

func (n *Network) Includes(ip net.IP) bool {
	for _, s := range n.subnets {
		if s.includes(ip) {
			return true
		}
	}
	return false
}

func (n *Network) GetPoolOfIP(ip net.IP) *Pool {
	for _, s := range n.subnets {
		for _, p := range s.pools {
			if p.includes(ip) {
				return p
			}
		}
	}
	return nil
}

func (n *Network) GetFreeLease(registered bool) (*store.Lease, *Pool) {
	for _, s := range n.subnets {
		if s.allowUnknown == registered {
			continue
		}
		for _, p := range s.pools {
			if l := p.getFreeLease(); l != nil {
				return l, p
			}
		}
	}
	return nil, nil
}

func (n *Network) GetFreeLeaseDesperate(registered bool) (*store.Lease, *Pool) {
	for _, s := range n.subnets {
		if s.allowUnknown == registered {
			continue
		}
		for _, p := range s.pools {
			if l := p.getFreeLeaseDesperate(); l != nil {
				return l, p
			}
		}
	}
	return nil, nil
}

func (n *Network) GetLeaseByMAC(mac net.HardwareAddr, registered bool) (*store.Lease, *Pool) {
	for _, s := range n.subnets {
		if s.allowUnknown == registered {
			continue
		}
		for _, p := range s.pools {
			p.m.RLock()
			for _, l := range p.leases {
				if bytes.Equal(l.MAC, mac) {
					p.m.RUnlock()
					return l, p
				}
			}
			p.m.RUnlock()
		}
	}
	return nil, nil
}

func (n *Network) GetLeaseByIP(ip net.IP, registered bool) (*store.Lease, *Pool) {
	for _, s := range n.subnets {
		if s.allowUnknown == registered {
			continue
		}
		for _, p := range s.pools {
			p.m.RLock()
			if l, ok := p.leases[ip.String()]; ok {
				p.m.RUnlock()
				return l, p
			}
			p.m.RUnlock()
		}
	}
	return nil, nil
}
