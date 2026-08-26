// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

type host struct {
	settings *settings
}

func newHost() *host {
	return &host{
		settings: newSettingsBlock(),
	}
}
