// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sconfig

import (
	"net"
	"time"
)

type Global struct {
	ServerIdentifier     net.IP
	settings             *Settings
	RegisteredSettings   *Settings
	regOptionsCached     bool
	UnregisteredSettings *Settings
	unregOptionsCached   bool
}

func newGlobal() *Global {
	g := &Global{
		settings:             newSettingsBlock(),
		RegisteredSettings:   newSettingsBlock(),
		UnregisteredSettings: newSettingsBlock(),
	}

	g.settings.DefaultLeaseTime = 3600 * time.Second
	g.settings.MaxLeaseTime = 3600 * time.Second
	g.settings.FreeLeaseAfter = 5 * time.Hour
	return g
}

// GetLeaseTime returns the lease time given the requested time req and if the client is registered.
// If req is 0 then the default lease time is returned. Otherwise it will return the lower of
// req and the maximum lease time. If a duration is not set for either, they will both be 1 week.
func (g *Global) getLeaseTime(req time.Duration, registered bool) time.Duration {
	if req <= 0 { // Default lease time
		if registered && g.RegisteredSettings.DefaultLeaseTime > 0 {
			return g.RegisteredSettings.DefaultLeaseTime
		}
		if !registered && g.UnregisteredSettings.DefaultLeaseTime > 0 {
			return g.UnregisteredSettings.DefaultLeaseTime
		}
		return g.settings.DefaultLeaseTime
	}

	// Client requested specific lease time
	if registered {
		// Client's request is less than or equal to max
		if g.RegisteredSettings.MaxLeaseTime > 0 {
			if req <= g.RegisteredSettings.MaxLeaseTime {
				return req
			}
			return g.RegisteredSettings.MaxLeaseTime
		}

		// Fallback to truly global settings
		if req <= g.settings.MaxLeaseTime {
			return req
		}
		return g.settings.MaxLeaseTime
	}

	// maxLeaseTime for unregistered
	// Client's request is less than or equal to max
	if g.UnregisteredSettings.MaxLeaseTime > 0 {
		if req <= g.UnregisteredSettings.MaxLeaseTime {
			return req
		}
		return g.UnregisteredSettings.MaxLeaseTime
	}

	// Fallback to truly global settings
	if req <= g.settings.MaxLeaseTime {
		return req
	}
	return g.settings.MaxLeaseTime
}

func (g *Global) getSettings(registered bool) *Settings {
	if registered && g.regOptionsCached {
		return g.RegisteredSettings
	} else if !registered && g.unregOptionsCached {
		return g.UnregisteredSettings
	}

	if registered {
		// Merge "global" settings into registered settings
		mergeSettings(g.RegisteredSettings, g.settings)
		g.regOptionsCached = true
		return g.RegisteredSettings
	}

	// Merge network "global" settings into unregistered settings
	mergeSettings(g.UnregisteredSettings, g.settings)
	g.unregOptionsCached = true
	return g.UnregisteredSettings
}
