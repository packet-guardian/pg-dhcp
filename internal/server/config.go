// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import "net"

// A Config is the parsed object generated from a PG-DHCP configuration file.
type Config struct {
	global   *global
	networks map[string]*network
	hosts    map[string]*hostOptions
}

func newConfig() *Config {
	return &Config{
		global:   newGlobal(),
		networks: make(map[string]*network),
		hosts:    make(map[string]*hostOptions),
	}
}

func (c *Config) searchNetworksFor(ip net.IP) *network {
	for _, network := range c.networks {
		if (ip.Equal(net.IPv4zero) && network.local) || network.includes(ip) {
			return network
		}
	}
	return nil
}

type hostOptions struct {
	settings     *settings
	reservations map[string]net.IP
}

func newHostOptions() *hostOptions {
	return &hostOptions{
		settings:     newSettingsBlock(),
		reservations: make(map[string]net.IP),
	}
}
