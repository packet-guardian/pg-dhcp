package sconfig

import (
	"bytes"
	"testing"
	"time"

	dhcp4 "github.com/packet-guardian/pg-dhcp/dhcp"
)

func TestSettingsMerge(t *testing.T) {
	d := newSettingsBlock()
	s := newSettingsBlock()

	d.Options[dhcp4.OptionBroadcastAddress] = []byte{10, 0, 254, 2}

	s.Options[dhcp4.OptionBroadcastAddress] = []byte{10, 0, 254, 3}
	s.Options[dhcp4.OptionDomainName] = []byte("example.com")
	s.DefaultLeaseTime = 360 * time.Second
	s.MaxLeaseTime = 500 * time.Second
	s.FreeLeaseAfter = 1800 * time.Second

	mergeSettings(d, s)

	if d.DefaultLeaseTime != s.DefaultLeaseTime {
		t.Errorf("Expected %d, got %d", s.DefaultLeaseTime, d.DefaultLeaseTime)
	}
	if d.MaxLeaseTime != s.MaxLeaseTime {
		t.Errorf("Expected %d, got %d", s.MaxLeaseTime, d.MaxLeaseTime)
	}
	if d.FreeLeaseAfter != s.FreeLeaseAfter {
		t.Errorf("Expected %d, got %d", s.FreeLeaseAfter, d.FreeLeaseAfter)
	}

	// Ensure the original value stays intact
	if bytes.Equal(d.Options[dhcp4.OptionBroadcastAddress], s.Options[dhcp4.OptionBroadcastAddress]) {
		t.Errorf("Expected %s, got %s", d.Options[dhcp4.OptionBroadcastAddress], s.Options[dhcp4.OptionBroadcastAddress])
	}

	// Ensure the new value is inherited
	if !bytes.Equal(d.Options[dhcp4.OptionDomainName], s.Options[dhcp4.OptionDomainName]) {
		t.Errorf("Expected %s, got %s", d.Options[dhcp4.OptionDomainName], s.Options[dhcp4.OptionDomainName])
	}
}
