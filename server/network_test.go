// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"net"
	"testing"
	"time"

	"github.com/lfkeitel/verbose"
	d4 "github.com/onesimus-systems/dhcp4"
	"github.com/packet-guardian/pg-dhcp/events"
	"github.com/packet-guardian/pg-dhcp/verification"
)

// TestGiveLeaseFromMultiplePools is targeted at the Network.getFreeLease()
// method. This test ensures that if a subnet has multiple pools and the first
// is already filled with claimed leases (not necessarily active leases), that
// it will go to the next pool in the subnet and get a lease from there.
// This test uses network3 in the test config and uses IP range 10.0.8.0/24
// with only an unregistered block.
func TestGiveLeaseFromMultiplePools(t *testing.T) {
	db, err := setUpLeaseStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownLeaseStore(db)

	sc := &ServerConfig{
		Verification: verification.NewNullVerifier(),
		Env:          EnvTesting,
		Log:          verbose.New(""),
		Store:        db,
		Events:       events.NewNullEmitter(),
	}

	// Setup Configuration
	c, err := ParseFile("./testdata/testConfig.conf")
	if err != nil {
		t.Fatalf("Test config failed parsing: %v", err)
	}

	network := c.networks["network3"]

	pool := network.subnets[0].pools[0]
	// Expire all leases, make one claimed
	for i := 0; i < pool.getCountOfIPs(); i++ {
		lease := pool.getFreeLease(sc)
		if lease == nil {
			t.Fatal("Pool returned nil lease")
		}
		lease.End = time.Now().Add(time.Duration(3610) * time.Second)
	}

	for _, l := range pool.leases {
		l.End = time.Now().Add(time.Duration(-1*c.global.unregisteredSettings.freeLeaseAfter) *
			time.Second).Add(time.Duration(300) * time.Second)
	}

	l := pool.leases["10.0.8.90"]
	l.End = time.Now().Add(time.Duration(-1*c.global.unregisteredSettings.freeLeaseAfter) * time.Second).Add(time.Duration(240) * time.Second)

	server := NewDHCPServer(c, sc)
	mac, _ := net.ParseMAC("12:34:56:12:34:56")

	opts := []d4.Option{
		d4.Option{
			Code:  d4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p := d4.RequestPacket(d4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.8.5"))

	// Process a DISCOVER request
	dp := server.ServeDHCP(p, d4.Discover, p.ParseOptions())
	if dp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(dp, []byte{0xa, 0x0, 0x8, 0x78}, t)
}
