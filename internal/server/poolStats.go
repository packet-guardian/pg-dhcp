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
	n := make([]string, len(c.Networks))
	i := 0
	for name := range c.Networks {
		n[i] = name
		i++
	}
	return n
}

func GetLeasesInNetwork(name string) []*models.Lease {
	net, ok := c.Networks[name]
	if !ok {
		return nil
	}
	return net.GetAllLeases()
}

func GetPoolStats() []*stats.PoolStat {
	poolStats := make([]*stats.PoolStat, 0)
	now := time.Now()
	regFreeTime := time.Duration(c.Global.RegisteredSettings.FreeLeaseAfter) * time.Second
	unRegFreeTime := time.Duration(c.Global.UnregisteredSettings.FreeLeaseAfter) * time.Second

	for _, n := range c.Networks {
		for _, s := range n.Subnets {
			for _, p := range s.Pools {
				ps := &stats.PoolStat{
					NetworkName: n.Name,
					Subnet:      s.Net.String(),
					Registered:  !s.AllowUnknown,
					Total:       p.GetCountOfIPs(),
					Start:       p.RangeStart.String(),
					End:         p.RangeEnd.String(),
				}

				for _, l := range p.Leases {
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
