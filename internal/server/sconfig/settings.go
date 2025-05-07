// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sconfig

import (
	"time"

	dhcp4 "github.com/packet-guardian/pg-dhcp/dhcp"
)

type Settings struct {
	Options          dhcp4.Options
	VendorOptions    dhcp4.Options
	DefaultLeaseTime time.Duration
	MaxLeaseTime     time.Duration
	FreeLeaseAfter   time.Duration
}

func newSettingsBlock() *Settings {
	return &Settings{
		Options:          make(dhcp4.Options),
		VendorOptions:    make(dhcp4.Options),
		FreeLeaseAfter:   0,
		DefaultLeaseTime: 0,
		MaxLeaseTime:     0,
	}
}

// mergeSettings will merge s into d.
func mergeSettings(d, s *Settings) {
	if d.DefaultLeaseTime == 0 {
		d.DefaultLeaseTime = s.DefaultLeaseTime
	}
	if d.MaxLeaseTime == 0 {
		d.MaxLeaseTime = s.MaxLeaseTime
	}
	if d.FreeLeaseAfter == 0 {
		d.FreeLeaseAfter = s.FreeLeaseAfter
	}

	for c, v := range s.Options {
		if _, ok := d.Options[c]; !ok {
			d.Options[c] = v
		}
	}
}

func (s *Settings) genVendorOption() []byte {
	length := 0

	for _, vd := range s.VendorOptions {
		length += 2 + len(vd) // 2 for code and data length
	}

	// I can't find in the RFCs exactly if this option
	// can be send over multiple CLV segments.
	if length == 0 || length > 255 {
		return nil
	}

	data := make([]byte, 0, length)
	for c, vd := range s.VendorOptions {
		vdlen := byte(len(vd))
		data = append(data, byte(c))
		data = append(data, vdlen)
		data = append(data, vd...)
	}

	return data
}
