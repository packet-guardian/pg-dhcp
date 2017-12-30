// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"time"

	"github.com/packet-guardian/pg-dhcp/models"
	"github.com/packet-guardian/pg-dhcp/stats"
)

func GetNetworkList() []string {
	n := make([]string, len(c.networks))
	i := 0
	for name := range c.networks {
		n[i] = name
		i++
	}
	return n
}

func GetLeasesInNetwork(name string) []*models.Lease {
	net, ok := c.networks[name]
	if !ok {
		return nil
	}
	return net.getAllLeases()
}

func GetPoolStats() []*stats.PoolStat {
	poolStats := make([]*stats.PoolStat, 0)
	now := time.Now()
	regFreeTime := time.Duration(c.global.registeredSettings.freeLeaseAfter) * time.Second
	unRegFreeTime := time.Duration(c.global.unregisteredSettings.freeLeaseAfter) * time.Second

	for _, n := range c.networks {
		for _, s := range n.subnets {
			for _, p := range s.pools {
				ps := &stats.PoolStat{
					NetworkName: n.name,
					Subnet:      s.net.String(),
					Registered:  !s.allowUnknown,
					Total:       p.getCountOfIPs(),
					Start:       p.rangeStart.String(),
					End:         p.rangeEnd.String(),
				}

				for _, l := range p.leases {
					if l.IsAbandoned {
						ps.Abandoned++
						continue
					}
					if !l.IsExpired() {
						ps.Active++
						continue
					}
					if !l.Registered && l.End.Add(unRegFreeTime).After(now) { // Unregisted lease expired
						ps.Claimed++
						continue
					}
					if l.Registered && l.End.Add(regFreeTime).After(now) { // Registered lease expired
						ps.Claimed++
						continue
					}
					if l.IsFree() {
						ps.Free++
						continue
					}
				}

				poolStats = append(poolStats, ps)
			}
		}
	}
	return poolStats
}
