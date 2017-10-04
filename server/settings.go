// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"time"

	"github.com/packet-guardian/pg-dhcp/dhcp"
)

type settings struct {
	options          map[dhcp4.OptionCode][]byte
	defaultLeaseTime time.Duration
	maxLeaseTime     time.Duration
	freeLeaseAfter   time.Duration
}

func newSettingsBlock() *settings {
	return &settings{
		options: make(map[dhcp4.OptionCode][]byte),
	}
}

// mergeSettings will merge s into d.
func mergeSettings(d, s *settings) {
	if d.defaultLeaseTime == 0 {
		d.defaultLeaseTime = s.defaultLeaseTime
	}
	if d.maxLeaseTime == 0 {
		d.maxLeaseTime = s.maxLeaseTime
	}
	if d.freeLeaseAfter == 0 {
		d.freeLeaseAfter = s.freeLeaseAfter
	}

	for c, v := range s.options {
		if _, ok := d.options[c]; !ok {
			d.options[c] = v
		}
	}
}
