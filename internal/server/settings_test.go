package server

import (
	"bytes"
	"testing"
	"time"

	dhcp4 "github.com/packet-guardian/pg-dhcp/dhcp"
)

func TestSettingsMerge(t *testing.T) {
	d := newSettingsBlock()
	s := newSettingsBlock()

	d.options[dhcp4.OptionBroadcastAddress] = []byte{10, 0, 254, 2}

	s.options[dhcp4.OptionBroadcastAddress] = []byte{10, 0, 254, 3}
	s.options[dhcp4.OptionDomainName] = []byte("example.com")
	s.defaultLeaseTime = 360 * time.Second
	s.maxLeaseTime = 500 * time.Second
	s.freeLeaseAfter = 1800 * time.Second

	mergeSettings(d, s)

	if d.defaultLeaseTime != s.defaultLeaseTime {
		t.Errorf("Expected %d, got %d", s.defaultLeaseTime, d.defaultLeaseTime)
	}
	if d.maxLeaseTime != s.maxLeaseTime {
		t.Errorf("Expected %d, got %d", s.maxLeaseTime, d.maxLeaseTime)
	}
	if d.freeLeaseAfter != s.freeLeaseAfter {
		t.Errorf("Expected %d, got %d", s.freeLeaseAfter, d.freeLeaseAfter)
	}

	// Ensure the original value stays intact
	if bytes.Equal(d.options[dhcp4.OptionBroadcastAddress], s.options[dhcp4.OptionBroadcastAddress]) {
		t.Errorf("Expected %s, got %s", d.options[dhcp4.OptionBroadcastAddress], s.options[dhcp4.OptionBroadcastAddress])
	}

	// Ensure the new value is inherited
	if !bytes.Equal(d.options[dhcp4.OptionDomainName], s.options[dhcp4.OptionDomainName]) {
		t.Errorf("Expected %s, got %s", d.options[dhcp4.OptionDomainName], s.options[dhcp4.OptionDomainName])
	}
}
