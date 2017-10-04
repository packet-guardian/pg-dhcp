// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import "github.com/packet-guardian/pg-dhcp/store"

// LoadLeases will import any current leases saved to the database.
func LoadLeases(s *store.Store, c *Config) error {
	s.ForEachLease(func(l *store.Lease) {
		// Check if the network exists
		n, ok := c.Networks[l.Network]
		if !ok {
			return
		}

		// Find the correct pool
		// TODO: Optimize this maybe with a temporary cache
	subnetLoop:
		for _, subnet := range n.subnets {
			if !subnet.includes(l.IP) {
				continue
			}

			for _, pool := range subnet.pools {
				if !pool.includes(l.IP) {
					continue
				}
				pool.leases[l.IP.String()] = l
				break subnetLoop
			}
		}
	})
	return nil
}
