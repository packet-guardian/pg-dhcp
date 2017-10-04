// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import (
	"net"
)

// A Config is the parsed object generated from a PG-DHCP configuration file.
type Config struct {
	Global   *Global
	Networks map[string]*Network
}

func newConfig() *Config {
	return &Config{
		Global:   newGlobal(),
		Networks: make(map[string]*Network),
	}
}

func (c *Config) SearchNetworksFor(ip net.IP) *Network {
	for _, network := range c.Networks {
		if network.Includes(ip) {
			return network
		}
	}
	return nil
}
