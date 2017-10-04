// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import (
	"net"
	"time"

	"github.com/packet-guardian/pg-dhcp/dhcp"
)

type subnet struct {
	allowUnknown  bool
	settings      *settings
	optionsCached bool
	net           *net.IPNet
	network       *Network
	pools         []*Pool
}

func newSubnet() *subnet {
	return &subnet{
		settings: newSettingsBlock(),
	}
}

// getLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If the subnet does not have an explicitly set duration for either,
// it will get the duration from its Network.
func (s *subnet) getLeaseTime(req time.Duration, registered bool) time.Duration {
	if req == 0 {
		if s.settings.defaultLeaseTime > 0 {
			return s.settings.defaultLeaseTime
		}
		// Save the result for later
		s.settings.defaultLeaseTime = s.network.getLeaseTime(req, registered)
		return s.settings.defaultLeaseTime
	}

	if s.settings.maxLeaseTime > 0 {
		if req <= s.settings.maxLeaseTime {
			return req
		}
		return s.settings.maxLeaseTime
	}

	// Save the result for later
	s.settings.maxLeaseTime = s.network.getLeaseTime(req, registered)

	if req <= s.settings.maxLeaseTime {
		return req
	}
	return s.settings.maxLeaseTime
}

func (s *subnet) getOptions(registered bool) dhcp4.Options {
	if s.optionsCached {
		return s.settings.options
	}

	mergeSettings(s.settings, s.network.getSettings(registered))
	s.optionsCached = true
	return s.settings.options
}

func (s *subnet) includes(ip net.IP) bool {
	return s.net.Contains(ip)
}
