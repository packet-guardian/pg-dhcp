// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sconfig

import (
	"net"
	"time"

	dhcp4 "github.com/packet-guardian/pg-dhcp/dhcp"
	"github.com/packet-guardian/pg-dhcp/models"
)

type Pool struct {
	RangeStart    net.IP
	RangeEnd      net.IP
	settings      *Settings
	optionsCached bool
	Leases        map[string]*models.Lease // IP -> Lease
	subnet        *Subnet
	nextFreeStart int
	ipsInPool     int
}

func newPool() *Pool {
	return &Pool{
		settings: newSettingsBlock(),
		Leases:   make(map[string]*models.Lease),
	}
}

func (p *Pool) GetCountOfIPs() int {
	if p.ipsInPool == 0 {
		p.ipsInPool = dhcp4.IPRange(p.RangeStart, p.RangeEnd)
	}
	return p.ipsInPool
}

// getLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If the pool does not have an explicitly set duration for either,
// it will get the duration from its subnet.
func (p *Pool) GetLeaseTime(req time.Duration, registered bool) time.Duration {
	if req == 0 {
		if p.settings.DefaultLeaseTime > 0 {
			return p.settings.DefaultLeaseTime
		}
		// Save the result for later
		p.settings.DefaultLeaseTime = p.subnet.getLeaseTime(req, registered)
		return p.settings.DefaultLeaseTime
	}

	if p.settings.MaxLeaseTime > 0 {
		if req < p.settings.MaxLeaseTime {
			return req
		}
		return p.settings.MaxLeaseTime
	}

	// Save the result for later
	p.settings.MaxLeaseTime = p.subnet.getLeaseTime(req, registered)

	if req <= p.settings.MaxLeaseTime {
		return req
	}
	return p.settings.MaxLeaseTime
}

func (p *Pool) GetOptions(registered bool) dhcp4.Options {
	if p.optionsCached {
		return p.settings.Options
	}

	higher := p.subnet.getOptions(registered)
	for c, v := range higher {
		if _, ok := p.settings.Options[c]; !ok {
			p.settings.Options[c] = v
		}
	}
	p.optionsCached = true
	return p.settings.Options
}

func (p *Pool) GetFreeLease() *models.Lease {
	now := time.Now()

	regFreeTime := p.subnet.network.global.RegisteredSettings.FreeLeaseAfter
	unRegFreeTime := p.subnet.network.global.UnregisteredSettings.FreeLeaseAfter
	// Find a candidate from the already used leases
	for _, l := range p.Leases {
		if l.IsAbandoned { // IP in use by a device we don't know about
			continue
		}
		if l.End.After(now) { // Active lease
			continue
		}
		if l.Offered && now.After(l.End) { // Lease was offered but not taken
			l.Offered = false
			return l
		}
		if !l.Registered && l.End.Add(unRegFreeTime).Before(now) { // Unregisted lease expired
			return l
		}
		if l.Registered && l.End.Add(regFreeTime).Before(now) { // Registered lease expired
			return l
		}
	}

	// No candidates, find the next available lease
	for i := p.nextFreeStart; i < p.GetCountOfIPs(); i++ {
		next := dhcp4.IPAdd(p.RangeStart, i)
		p.nextFreeStart = i + 1

		// Check if IP has a lease
		// Sanity check
		_, ok := p.Leases[next.String()]
		if ok {
			continue
		}

		// IP has no lease with it, no lock since this is a new object
		// and guarenteed to not be anywhere else yet.
		l := models.NewLease()
		l.IP = next
		l.Network = p.subnet.network.Name
		l.Registered = !p.subnet.AllowUnknown
		p.Leases[next.String()] = l
		return l
	}

	// We've exhausted all possibilities, admit defeat.
	return nil
}

func (p *Pool) getFreeLeaseDesperate() *models.Lease {
	now := time.Now()

	// No free leases, bring out the big guns
	// Find the oldest expired lease
	var longestExpiredLease *models.Lease
	for _, l := range p.Leases {
		if l.End.After(now) { // Skip active leases
			continue
		}

		if longestExpiredLease == nil {
			longestExpiredLease = l
			continue
		}

		if l.End.Before(longestExpiredLease.End) {
			longestExpiredLease = l
		}
	}

	if longestExpiredLease != nil {
		return longestExpiredLease
	}

	// Now we're getting desperate
	// Check abandoned leases for availability
	for _, l := range p.Leases {
		if l.IsAbandoned { // Skip non-abandoned leases
			l.IsAbandoned = false
			return l
		}
	}
	return nil
}

func (p *Pool) Includes(ip net.IP) bool {
	return dhcp4.IPInRange(p.RangeStart, p.RangeEnd, ip)
}
