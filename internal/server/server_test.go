// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/lfkeitel/verbose"
	d4 "github.com/packet-guardian/pg-dhcp/dhcp"
	"github.com/packet-guardian/pg-dhcp/models"
	"github.com/packet-guardian/pg-dhcp/store"
)

type fatalLogger interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

func setUpTest1(t fatalLogger) *Handler {
	db, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}

	// Setup Configuration
	c, err := ParseFile("./testdata/testConfig.conf")
	if err != nil {
		t.Fatalf("Test config failed parsing: %v", err)
	}

	sc := &ServerConfig{
		Env:   EnvTesting,
		Log:   verbose.New(""),
		Store: db,
	}

	return NewDHCPServer(c, sc)
}

func tearDownTest1(h *Handler) {
	tearDownStore(h.c.Store)
}

func setDevice(s store.Store, m net.HardwareAddr, r, b bool) {
	s.PutDevice(&models.Device{
		MAC:         m,
		Registered:  r,
		Blacklisted: b,
	})
}

func TestDiscover(t *testing.T) {
	server := setUpTest1(t)
	defer tearDownTest1(server)
	mac, _ := net.ParseMAC("12:34:56:12:34:56")

	// Round 1 - Test Registered Device
	setDevice(server.c.Store, mac, true, false)

	// Create test request packet
	opts := []d4.Option{
		d4.Option{
			Code:  d4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p := d4.RequestPacket(d4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a DISCOVER request
	start := time.Now()
	dp := server.ServeDHCP(p, d4.Discover, p.ParseOptions())
	t.Logf("Discover took: %v", time.Since(start))

	if dp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(dp, []byte{0xa, 0x0, 0x2, 0xa}, t)
	options := checkOptions(dp, d4.Options{
		d4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		d4.OptionRouter:             []byte{0xa, 0x0, 0x2, 0x1},
		d4.OptionDomainNameServer:   []byte{0xa, 0x1, 0x0, 0x1, 0xa, 0x1, 0x0, 0x2},
		d4.OptionDomainName:         []byte("example.com"),
		d4.OptionIPAddressLeaseTime: []byte{0x0, 0x1, 0x51, 0x80},
	}, t)

	opts = []d4.Option{
		d4.Option{
			Code:  d4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
		d4.Option{
			Code:  d4.OptionServerIdentifier,
			Value: []byte(options[d4.OptionServerIdentifier]),
		},
		d4.Option{
			Code:  d4.OptionRequestedIPAddress,
			Value: []byte(dp.YIAddr().To4()),
		},
	}
	p = d4.RequestPacket(d4.Request, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a REQUEST request
	start = time.Now()
	rp := server.ServeDHCP(p, d4.Request, p.ParseOptions())
	t.Logf("Request took: %v", time.Since(start))

	if rp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(rp, dp.YIAddr(), t)
	checkOptions(rp, d4.Options{
		d4.OptionDHCPMessageType:    []byte{0x5},
		d4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		d4.OptionRouter:             []byte{0xa, 0x0, 0x2, 0x1},
		d4.OptionDomainNameServer:   []byte{0xa, 0x1, 0x0, 0x1, 0xa, 0x1, 0x0, 0x2},
		d4.OptionDomainName:         []byte("example.com"),
		d4.OptionIPAddressLeaseTime: []byte{0x0, 0x1, 0x51, 0x80},
	}, t)

	// ROUND 2 - Fight! Test Unregistered Device
	setDevice(server.c.Store, mac, false, false)

	opts = []d4.Option{
		d4.Option{
			Code:  d4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p = d4.RequestPacket(d4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a DISCOVER request
	start = time.Now()
	dp = server.ServeDHCP(p, d4.Discover, p.ParseOptions())
	t.Logf("Discover took: %v", time.Since(start))

	if dp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(dp, []byte{0xa, 0x0, 0x1, 0xa}, t)
	checkOptions(dp, d4.Options{
		d4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		d4.OptionRouter:             []byte{0xa, 0x0, 0x1, 0x1},
		d4.OptionDomainNameServer:   []byte{0xa, 0x0, 0x0, 0x1},
		d4.OptionDomainName:         []byte("example.com"),
		d4.OptionIPAddressLeaseTime: []byte{0x0, 0x0, 0x1, 0x68},
	}, t)

	opts = []d4.Option{
		d4.Option{
			Code:  d4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
		d4.Option{
			Code:  d4.OptionServerIdentifier,
			Value: []byte(options[d4.OptionServerIdentifier]),
		},
		d4.Option{
			Code:  d4.OptionRequestedIPAddress,
			Value: []byte(dp.YIAddr().To4()),
		},
	}
	p = d4.RequestPacket(d4.Request, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a REQUEST request
	start = time.Now()
	rp = server.ServeDHCP(p, d4.Request, p.ParseOptions())
	t.Logf("Request took: %v", time.Since(start))

	if rp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(rp, dp.YIAddr(), t)
	checkOptions(rp, d4.Options{
		d4.OptionDHCPMessageType:    []byte{0x5},
		d4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		d4.OptionRouter:             []byte{0xa, 0x0, 0x1, 0x1},
		d4.OptionDomainNameServer:   []byte{0xa, 0x0, 0x0, 0x1},
		d4.OptionDomainName:         []byte("example.com"),
		d4.OptionIPAddressLeaseTime: []byte{0x0, 0x0, 0x1, 0x68},
	}, t)
}

func TestBlockBlacklisted(t *testing.T) {
	server := setUpTest1(t)
	defer tearDownTest1(server)
	mac, _ := net.ParseMAC("12:34:56:12:34:56")

	// Round 1 - Test Registered Device
	setDevice(server.c.Store, mac, true, true)
	server.c.BlockBlacklist = true

	// Create test request packet
	opts := []d4.Option{
		d4.Option{
			Code:  d4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p := d4.RequestPacket(d4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a DISCOVER request
	start := time.Now()
	dp := server.ServeDHCP(p, d4.Discover, p.ParseOptions())
	t.Logf("Discover took: %v", time.Since(start))

	if dp != nil {
		t.Fatal("Blacklisted devices received a reply instead of being blocked.")
	}
}

func checkIP(p d4.Packet, expected net.IP, t *testing.T) {
	if !bytes.Equal(p.YIAddr().To4(), expected.To4()) {
		t.Errorf("Incorrect IP. Expected %v, got %v", expected, p.YIAddr())
	}
}

func checkOptions(p d4.Packet, ops d4.Options, t *testing.T) d4.Options {
	options := p.ParseOptions()
	for o, v := range ops {
		if val, ok := options[o]; !ok { // 0x23 (51)
			t.Errorf("%s not received", o.String())
		} else if !bytes.Equal(val, v) {
			t.Errorf("Incorrect %s. Expected %v, got %v", o.String(), v, val)
		}
	}
	return options
}

func BenchmarkDHCPDiscover(b *testing.B) {
	server := setUpTest1(b)
	defer tearDownTest1(server)

	mac, _ := net.ParseMAC("12:34:56:12:34:56")
	server.c.Store.PutDevice(&models.Device{
		MAC:         mac,
		Registered:  true,
		Blacklisted: false,
	})

	pool := c.networks["network1"].subnets[1].pools[0] // Registered pool

	// Create test request packet
	opts := []d4.Option{
		d4.Option{
			Code:  d4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p := d4.RequestPacket(d4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))
	unixZero := time.Unix(0, 0)

	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		dp := server.ServeDHCP(p, d4.Discover, p.ParseOptions())
		b.StopTimer()

		if dp == nil {
			b.Fatal("ServeDHCP returned nil")
		}
		pool.leases["10.0.2.10"].End = unixZero
	}
}
