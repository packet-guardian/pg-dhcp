// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/lfkeitel/verbose"
	d4 "github.com/onesimus-systems/dhcp4"
)

func setUpTest1(t *testing.T) (*Handler, *testDeviceStore, *testLeaseStore) {
	// Setup Confuration
	c, err := ParseFile("./testdata/testConfig.conf")
	if err != nil {
		t.Fatalf("Test config failed parsing: %v", err)
	}

	ds := &testDeviceStore{}
	ls := &testLeaseStore{}

	return NewDHCPServer(c, &ServerConfig{
		LeaseStore:  ls,
		DeviceStore: ds,
		Env:         EnvTesting,
		LogPath:     "",
		Log:         verbose.New(""),
	}), ds, ls
}

func TestDiscover(t *testing.T) {
	server, ds, _ := setUpTest1(t)
	mac, _ := net.ParseMAC("12:34:56:12:34:56")

	regTestDevice := &testDevice{
		store:      ds,
		registered: true,
		mac:        mac,
	}

	unregTestDevice := &testDevice{
		store:      ds,
		registered: false,
		mac:        mac,
	}

	// Round 1 - Test Registered Device
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
	ds.setNextDevice(regTestDevice)
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
	ds.setNextDevice(regTestDevice)
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
	opts = []d4.Option{
		d4.Option{
			Code:  d4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p = d4.RequestPacket(d4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a DISCOVER request
	ds.setNextDevice(unregTestDevice)
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
	ds.setNextDevice(unregTestDevice)
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
	// Setup Confuration
	c, err := ParseFile("./testdata/testConfig.conf")
	if err != nil {
		b.Fatalf("Test config failed parsing: %v", err)
	}

	ds := &testDeviceStore{}
	ls := &testLeaseStore{}

	server := NewDHCPServer(c, &ServerConfig{
		LeaseStore:  ls,
		DeviceStore: ds,
		Env:         EnvTesting,
		LogPath:     "",
	})

	mac, _ := net.ParseMAC("12:34:56:12:34:56")

	unregTestDevice := &testDevice{
		store:      ds,
		registered: true,
		mac:        mac,
	}

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

	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		ds.setNextDevice(unregTestDevice)

		b.StartTimer()
		dp := server.ServeDHCP(p, d4.Discover, p.ParseOptions())
		b.StopTimer()

		if dp == nil {
			b.Fatal("ServeDHCP returned nil")
		}
		pool.leases["10.0.2.10"].End = unixZero
	}
}
