// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sconfig

type Host struct {
	Settings *Settings
}

func newHost() *Host {
	return &Host{
		Settings: newSettingsBlock(),
	}
}
