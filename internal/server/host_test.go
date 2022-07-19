// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"bytes"
	"testing"

	dhcp4 "github.com/packet-guardian/pg-dhcp/dhcp"
)

func TestHostConfig(t *testing.T) {
	// Setup Configuration
	c, err := ParseFile("./testdata/testConfig.conf")
	if err != nil {
		t.Fatalf("Test config failed parsing: %v", err)
	}

	if len(c.hosts) != 2 {
		t.Errorf("Incorrect number of hosts. Expected 2, got %d", len(c.hosts))
	}

	host1, exists := c.hosts["12:34:56:ab:cd:ef"]
	if !exists {
		t.Fatal("Host '12:34:56:ab:cd:ef' doesn't exist in config")
	}
	if len(host1.settings.options) != 1 {
		t.Errorf("Incorrect number of host1 options. Expected 1, got %d", len(host1.settings.options))
	}
	if !bytes.Equal([]byte{192, 168, 0, 10}, host1.settings.options[dhcp4.OptionDomainNameServer]) {
		t.Errorf("Expected %v, got %v", []byte{192, 168, 0, 10}, host1.settings.options[dhcp4.OptionDomainNameServer])
	}

	host2, exists := c.hosts["12:34:56:78:cd:ef"]
	if !exists {
		t.Fatal("Host '12:34:56:78:cd:ef' doesn't exist in config")
	}
	if len(host2.settings.options) != 2 {
		t.Errorf("Incorrect number of host2 options. Expected 1, got %d", len(host2.settings.options))
	}
	if !bytes.Equal([]byte{192, 168, 0, 11}, host2.settings.options[dhcp4.OptionDomainNameServer]) {
		t.Errorf("Expected %v, got %v", []byte{192, 168, 0, 11}, host2.settings.options[dhcp4.OptionDomainNameServer])
	}

	testBytes := []byte("This is some text")
	if !bytes.Equal(testBytes, host2.settings.options[125]) {
		t.Errorf("Expected %v, got %v", testBytes, host2.settings.options[125])
	}
}
