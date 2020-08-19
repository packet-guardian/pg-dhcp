// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import "testing"

func TestParseConfig(t *testing.T) {
	// TODO: Actually check the underlying config to make sure it matches the parsed config
	_, err := ParseFile("./testdata/testConfig.conf")
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseIncludedConfigs(t *testing.T) {
	c, err := ParseFile("./testdata/includeConfig.conf")
	if err != nil {
		t.Fatal(err)
	}

	// Three networks means it processed the include correctly and it continued correctly
	if len(c.networks) != 3 {
		t.Fatalf("Incorrect number of networks. Expected 3, got %d", len(c.networks))
	}
}

func TestParseHostOptions(t *testing.T) {
	c, err := ParseFile("./testdata/hostOptions.conf")
	if err != nil {
		t.Fatal(err)
	}

	// Three networks means it processed the include correctly and it continued correctly
	if len(c.hosts) != 1 {
		t.Fatalf("Incorrect number of hosts. Expected 1, got %d", len(c.hosts))
	}
}
