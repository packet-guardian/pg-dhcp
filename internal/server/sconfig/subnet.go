// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sconfig

import (
	"net"
	"time"

	dhcp4 "github.com/packet-guardian/pg-dhcp/dhcp"
)

type Subnet struct {
	AllowUnknown  bool
	settings      *Settings
	optionsCached bool
	Net           *net.IPNet
	network       *Network
	Pools         []*Pool
}

func newSubnet() *Subnet {
	return &Subnet{
		settings: newSettingsBlock(),
	}
}

// getLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If the subnet does not have an explicitly set duration for either,
// it will get the duration from its Network.
func (s *Subnet) getLeaseTime(req time.Duration, registered bool) time.Duration {
	if req == 0 {
		if s.settings.DefaultLeaseTime > 0 {
			return s.settings.DefaultLeaseTime
		}
		// Save the result for later
		s.settings.DefaultLeaseTime = s.network.getLeaseTime(req, registered)
		return s.settings.DefaultLeaseTime
	}

	if s.settings.MaxLeaseTime > 0 {
		if req <= s.settings.MaxLeaseTime {
			return req
		}
		return s.settings.MaxLeaseTime
	}

	// Save the result for later
	s.settings.MaxLeaseTime = s.network.getLeaseTime(req, registered)

	if req <= s.settings.MaxLeaseTime {
		return req
	}
	return s.settings.MaxLeaseTime
}

func (s *Subnet) getOptions(registered bool) dhcp4.Options {
	if s.optionsCached {
		return s.settings.Options
	}

	mergeSettings(s.settings, s.network.getSettings(registered))
	s.optionsCached = true
	return s.settings.Options
}

func (s *Subnet) Includes(ip net.IP) bool {
	return s.Net.Contains(ip)
}
