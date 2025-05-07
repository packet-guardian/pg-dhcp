// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/lfkeitel/verbose/v5"
	dhcp4 "github.com/packet-guardian/pg-dhcp/dhcp"
	"github.com/packet-guardian/pg-dhcp/internal/server/sconfig"
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
	c, err := sconfig.ParseFile("./testdata/testConfig.conf")
	if err != nil {
		t.Fatalf("Test config failed parsing: %v", err)
	}

	sc := &ServerConfig{
		Env:      EnvTesting,
		Log:      verbose.New(),
		Store:    db,
		Networks: c,
	}

	return NewDHCPServer(sc)
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
	opts := []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23, 43},
		},
	}
	p := dhcp4.RequestPacket(dhcp4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a DISCOVER request
	start := time.Now()
	dp := server.ServeDHCP(p, dhcp4.Discover, p.ParseOptions())
	t.Logf("Discover took: %v", time.Since(start))

	if dp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(dp, []byte{0xa, 0x0, 0x2, 0xa}, t)
	options := checkOptions(dp, dhcp4.Options{
		dhcp4.OptionSubnetMask:                []byte{0xff, 0xff, 0xff, 0x0},
		dhcp4.OptionRouter:                    []byte{0xa, 0x0, 0x2, 0x1},
		dhcp4.OptionDomainNameServer:          []byte{0xa, 0x1, 0x0, 0x1, 0xa, 0x1, 0x0, 0x2},
		dhcp4.OptionDomainName:                []byte("example.com"),
		dhcp4.OptionIPAddressLeaseTime:        []byte{0x0, 0x1, 0x51, 0x80},
		dhcp4.OptionVendorSpecificInformation: []byte{18, 4, 't', 'e', 's', 't'},
	}, t)

	opts = []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23, 43},
		},
		dhcp4.Option{
			Code:  dhcp4.OptionServerIdentifier,
			Value: []byte(options[dhcp4.OptionServerIdentifier]),
		},
		dhcp4.Option{
			Code:  dhcp4.OptionRequestedIPAddress,
			Value: []byte(dp.YIAddr().To4()),
		},
	}
	p = dhcp4.RequestPacket(dhcp4.Request, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a REQUEST request
	start = time.Now()
	rp := server.ServeDHCP(p, dhcp4.Request, p.ParseOptions())
	t.Logf("Request took: %v", time.Since(start))

	if rp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(rp, dp.YIAddr(), t)
	checkOptions(rp, dhcp4.Options{
		dhcp4.OptionDHCPMessageType:           []byte{0x5},
		dhcp4.OptionSubnetMask:                []byte{0xff, 0xff, 0xff, 0x0},
		dhcp4.OptionRouter:                    []byte{0xa, 0x0, 0x2, 0x1},
		dhcp4.OptionDomainNameServer:          []byte{0xa, 0x1, 0x0, 0x1, 0xa, 0x1, 0x0, 0x2},
		dhcp4.OptionDomainName:                []byte("example.com"),
		dhcp4.OptionIPAddressLeaseTime:        []byte{0x0, 0x1, 0x51, 0x80},
		dhcp4.OptionVendorSpecificInformation: []byte{18, 4, 't', 'e', 's', 't'},
	}, t)

	// ROUND 2 - Fight! Test Unregistered Device
	setDevice(server.c.Store, mac, false, false)

	opts = []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p = dhcp4.RequestPacket(dhcp4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a DISCOVER request
	start = time.Now()
	dp = server.ServeDHCP(p, dhcp4.Discover, p.ParseOptions())
	t.Logf("Discover took: %v", time.Since(start))

	if dp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(dp, []byte{0xa, 0x0, 0x1, 0xa}, t)
	checkOptions(dp, dhcp4.Options{
		dhcp4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		dhcp4.OptionRouter:             []byte{0xa, 0x0, 0x1, 0x1},
		dhcp4.OptionDomainNameServer:   []byte{0xa, 0x0, 0x0, 0x1},
		dhcp4.OptionDomainName:         []byte("example.com"),
		dhcp4.OptionIPAddressLeaseTime: []byte{0x0, 0x0, 0x1, 0x68},
	}, t)

	opts = []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
		dhcp4.Option{
			Code:  dhcp4.OptionServerIdentifier,
			Value: []byte(options[dhcp4.OptionServerIdentifier]),
		},
		dhcp4.Option{
			Code:  dhcp4.OptionRequestedIPAddress,
			Value: []byte(dp.YIAddr().To4()),
		},
	}
	p = dhcp4.RequestPacket(dhcp4.Request, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a REQUEST request
	start = time.Now()
	rp = server.ServeDHCP(p, dhcp4.Request, p.ParseOptions())
	t.Logf("Request took: %v", time.Since(start))

	if rp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(rp, dp.YIAddr(), t)
	checkOptions(rp, dhcp4.Options{
		dhcp4.OptionDHCPMessageType:    []byte{0x5},
		dhcp4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		dhcp4.OptionRouter:             []byte{0xa, 0x0, 0x1, 0x1},
		dhcp4.OptionDomainNameServer:   []byte{0xa, 0x0, 0x0, 0x1},
		dhcp4.OptionDomainName:         []byte("example.com"),
		dhcp4.OptionIPAddressLeaseTime: []byte{0x0, 0x0, 0x1, 0x68},
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
	opts := []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p := dhcp4.RequestPacket(dhcp4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))

	// Process a DISCOVER request
	start := time.Now()
	dp := server.ServeDHCP(p, dhcp4.Discover, p.ParseOptions())
	t.Logf("Discover took: %v", time.Since(start))

	if dp != nil {
		t.Fatal("Blacklisted devices received a reply instead of being blocked.")
	}
}

func TestIgnoreRegistration(t *testing.T) {
	server := setUpTest1(t)
	defer tearDownTest1(server)
	mac, _ := net.ParseMAC("12:34:56:12:34:56")

	// Round 1 - Test Registered Device
	setDevice(server.c.Store, mac, true, false)

	// Create test request packet
	opts := []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p := dhcp4.RequestPacket(dhcp4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.10.5"))

	// Process a DISCOVER request
	start := time.Now()
	dp := server.ServeDHCP(p, dhcp4.Discover, p.ParseOptions())
	t.Logf("Discover took: %v", time.Since(start))

	if dp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(dp, []byte{0xa, 0x0, 0xa, 0xa}, t)
	options := checkOptions(dp, dhcp4.Options{
		dhcp4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		dhcp4.OptionRouter:             []byte{0xa, 0x0, 0xa, 0x1},
		dhcp4.OptionDomainNameServer:   []byte{0xa, 0x0, 0x0, 0x1},
		dhcp4.OptionDomainName:         []byte("example.com"),
		dhcp4.OptionIPAddressLeaseTime: []byte{0x0, 0x0, 0x1, 0x68},
	}, t)

	opts = []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
		dhcp4.Option{
			Code:  dhcp4.OptionServerIdentifier,
			Value: []byte(options[dhcp4.OptionServerIdentifier]),
		},
		dhcp4.Option{
			Code:  dhcp4.OptionRequestedIPAddress,
			Value: []byte(dp.YIAddr().To4()),
		},
	}
	p = dhcp4.RequestPacket(dhcp4.Request, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.10.5"))

	// Process a REQUEST request
	start = time.Now()
	rp := server.ServeDHCP(p, dhcp4.Request, p.ParseOptions())
	t.Logf("Request took: %v", time.Since(start))

	if rp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(rp, dp.YIAddr(), t)
	checkOptions(rp, dhcp4.Options{
		dhcp4.OptionDHCPMessageType:    []byte{0x5},
		dhcp4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		dhcp4.OptionRouter:             []byte{0xa, 0x0, 0xa, 0x1},
		dhcp4.OptionDomainNameServer:   []byte{0xa, 0x0, 0x0, 0x1},
		dhcp4.OptionDomainName:         []byte("example.com"),
		dhcp4.OptionIPAddressLeaseTime: []byte{0x0, 0x0, 0x1, 0x68},
	}, t)

	// ROUND 2 - Fight! Test Unregistered Device, should be exactly the same
	setDevice(server.c.Store, mac, false, false)

	opts = []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p = dhcp4.RequestPacket(dhcp4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.10.5"))

	// Process a DISCOVER request
	start = time.Now()
	dp = server.ServeDHCP(p, dhcp4.Discover, p.ParseOptions())
	t.Logf("Discover took: %v", time.Since(start))

	if dp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(dp, []byte{0xa, 0x0, 0xa, 0xa}, t)
	options = checkOptions(dp, dhcp4.Options{
		dhcp4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		dhcp4.OptionRouter:             []byte{0xa, 0x0, 0xa, 0x1},
		dhcp4.OptionDomainNameServer:   []byte{0xa, 0x0, 0x0, 0x1},
		dhcp4.OptionDomainName:         []byte("example.com"),
		dhcp4.OptionIPAddressLeaseTime: []byte{0x0, 0x0, 0x1, 0x68},
	}, t)

	opts = []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
		dhcp4.Option{
			Code:  dhcp4.OptionServerIdentifier,
			Value: []byte(options[dhcp4.OptionServerIdentifier]),
		},
		dhcp4.Option{
			Code:  dhcp4.OptionRequestedIPAddress,
			Value: []byte(dp.YIAddr().To4()),
		},
	}
	p = dhcp4.RequestPacket(dhcp4.Request, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.10.5"))

	// Process a REQUEST request
	start = time.Now()
	rp = server.ServeDHCP(p, dhcp4.Request, p.ParseOptions())
	t.Logf("Request took: %v", time.Since(start))

	if rp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(rp, dp.YIAddr(), t)
	checkOptions(rp, dhcp4.Options{
		dhcp4.OptionDHCPMessageType:    []byte{0x5},
		dhcp4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		dhcp4.OptionRouter:             []byte{0xa, 0x0, 0xa, 0x1},
		dhcp4.OptionDomainNameServer:   []byte{0xa, 0x0, 0x0, 0x1},
		dhcp4.OptionDomainName:         []byte("example.com"),
		dhcp4.OptionIPAddressLeaseTime: []byte{0x0, 0x0, 0x1, 0x68},
	}, t)
}

func TestHostOptionsHost1(t *testing.T) {
	server := setUpTest1(t)
	defer tearDownTest1(server)
	mac, _ := net.ParseMAC("12:34:56:ab:cd:ef")

	// Round 1 - Test Registered Device
	setDevice(server.c.Store, mac, true, false)

	// Create test request packet
	opts := []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p := dhcp4.RequestPacket(dhcp4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.10.5"))

	// Process a DISCOVER request
	start := time.Now()
	dp := server.ServeDHCP(p, dhcp4.Discover, p.ParseOptions())
	t.Logf("Discover took: %v", time.Since(start))

	if dp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(dp, []byte{0xa, 0x0, 0xa, 0xa}, t)
	options := checkOptions(dp, dhcp4.Options{
		dhcp4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		dhcp4.OptionRouter:             []byte{0xa, 0x0, 0xa, 0x1},
		dhcp4.OptionDomainNameServer:   []byte{192, 168, 0, 10},
		dhcp4.OptionDomainName:         []byte("example.com"),
		dhcp4.OptionIPAddressLeaseTime: []byte{0x0, 0x0, 0x1, 0x68},
	}, t)

	opts = []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
		dhcp4.Option{
			Code:  dhcp4.OptionServerIdentifier,
			Value: []byte(options[dhcp4.OptionServerIdentifier]),
		},
		dhcp4.Option{
			Code:  dhcp4.OptionRequestedIPAddress,
			Value: []byte(dp.YIAddr().To4()),
		},
	}
	p = dhcp4.RequestPacket(dhcp4.Request, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.10.5"))

	// Process a REQUEST request
	start = time.Now()
	rp := server.ServeDHCP(p, dhcp4.Request, p.ParseOptions())
	t.Logf("Request took: %v", time.Since(start))

	if rp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(rp, dp.YIAddr(), t)
	checkOptions(rp, dhcp4.Options{
		dhcp4.OptionDHCPMessageType:    []byte{0x5},
		dhcp4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		dhcp4.OptionRouter:             []byte{0xa, 0x0, 0xa, 0x1},
		dhcp4.OptionDomainNameServer:   []byte{192, 168, 0, 10},
		dhcp4.OptionDomainName:         []byte("example.com"),
		dhcp4.OptionIPAddressLeaseTime: []byte{0x0, 0x0, 0x1, 0x68},
	}, t)
}

func TestHostOptionsHost2(t *testing.T) {
	server := setUpTest1(t)
	defer tearDownTest1(server)
	mac, _ := net.ParseMAC("12:34:56:78:cd:ef")

	// Round 1 - Test Registered Device
	setDevice(server.c.Store, mac, true, false)

	// Create test request packet
	opts := []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23, 125},
		},
	}
	p := dhcp4.RequestPacket(dhcp4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.10.5"))

	// Process a DISCOVER request
	start := time.Now()
	dp := server.ServeDHCP(p, dhcp4.Discover, p.ParseOptions())
	t.Logf("Discover took: %v", time.Since(start))

	if dp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(dp, []byte{0xa, 0x0, 0xa, 0xa}, t)
	options := checkOptions(dp, dhcp4.Options{
		dhcp4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		dhcp4.OptionRouter:             []byte{0xa, 0x0, 0xa, 0x1},
		dhcp4.OptionDomainNameServer:   []byte{192, 168, 0, 11},
		dhcp4.OptionDomainName:         []byte("example.com"),
		dhcp4.OptionIPAddressLeaseTime: []byte{0x0, 0x0, 0x1, 0x68},
		125:                            []byte("This is some text"),
	}, t)

	opts = []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23, 125},
		},
		dhcp4.Option{
			Code:  dhcp4.OptionServerIdentifier,
			Value: []byte(options[dhcp4.OptionServerIdentifier]),
		},
		dhcp4.Option{
			Code:  dhcp4.OptionRequestedIPAddress,
			Value: []byte(dp.YIAddr().To4()),
		},
	}
	p = dhcp4.RequestPacket(dhcp4.Request, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.10.5"))

	// Process a REQUEST request
	start = time.Now()
	rp := server.ServeDHCP(p, dhcp4.Request, p.ParseOptions())
	t.Logf("Request took: %v", time.Since(start))

	if rp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(rp, dp.YIAddr(), t)
	checkOptions(rp, dhcp4.Options{
		dhcp4.OptionDHCPMessageType:    []byte{0x5},
		dhcp4.OptionSubnetMask:         []byte{0xff, 0xff, 0xff, 0x0},
		dhcp4.OptionRouter:             []byte{0xa, 0x0, 0xa, 0x1},
		dhcp4.OptionDomainNameServer:   []byte{192, 168, 0, 11},
		dhcp4.OptionDomainName:         []byte("example.com"),
		dhcp4.OptionIPAddressLeaseTime: []byte{0x0, 0x0, 0x1, 0x68},
		125:                            []byte("This is some text"),
	}, t)
}

func checkIP(p dhcp4.Packet, expected net.IP, t *testing.T) {
	if !bytes.Equal(p.YIAddr().To4(), expected.To4()) {
		t.Errorf("Incorrect IP. Expected %v, got %v", expected, p.YIAddr())
	}
}

func checkOptions(p dhcp4.Packet, ops dhcp4.Options, t *testing.T) dhcp4.Options {
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

	pool := c.Networks["network1"].Subnets[1].Pools[0] // Registered pool

	// Create test request packet
	opts := []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p := dhcp4.RequestPacket(dhcp4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.1.5"))
	unixZero := time.Unix(0, 0)

	b.StopTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		dp := server.ServeDHCP(p, dhcp4.Discover, p.ParseOptions())
		b.StopTimer()

		if dp == nil {
			b.Fatal("ServeDHCP returned nil")
		}
		pool.Leases["10.0.2.10"].End = unixZero
	}
}

// TestGiveLeaseFromMultiplePools is targeted at the Network.getFreeLease()
// method. This test ensures that if a subnet has multiple pools and the first
// is already filled with claimed leases (not necessarily active leases), that
// it will go to the next pool in the subnet and get a lease from there.
// This test uses network3 in the test config and uses IP range 10.0.8.0/24
// with only an unregistered block.
func TestGiveLeaseFromMultiplePools(t *testing.T) {
	db, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(db)

	// Setup Configuration
	c, err := sconfig.ParseFile("./testdata/testConfig.conf")
	if err != nil {
		t.Fatalf("Test config failed parsing: %v", err)
	}

	network := c.Networks["network3"]

	pool := network.Subnets[0].Pools[0]
	// Expire all leases, make one claimed
	for i := 0; i < pool.GetCountOfIPs(); i++ {
		lease := pool.GetFreeLease()
		if lease == nil {
			t.Fatal("Pool returned nil lease")
		}
		lease.End = time.Now().Add(time.Duration(3610) * time.Second)
	}

	for _, l := range pool.Leases {
		l.End = time.Now().Add(time.Duration(-1*c.Global.UnregisteredSettings.FreeLeaseAfter) *
			time.Second).Add(time.Duration(300) * time.Second)
	}

	l := pool.Leases["10.0.8.90"]
	l.End = time.Now().Add(time.Duration(-1*c.Global.UnregisteredSettings.FreeLeaseAfter) * time.Second).Add(time.Duration(240) * time.Second)

	sc := &ServerConfig{
		Env:      EnvTesting,
		Log:      verbose.New(),
		Store:    db,
		Networks: c,
	}

	server := NewDHCPServer(sc)
	mac, _ := net.ParseMAC("12:34:56:12:34:56")

	opts := []dhcp4.Option{
		dhcp4.Option{
			Code:  dhcp4.OptionParameterRequestList,
			Value: []byte{0x1, 0x3, 0x6, 0xf, 0x23},
		},
	}
	p := dhcp4.RequestPacket(dhcp4.Discover, mac, nil, nil, false, opts)
	p.SetGIAddr(net.ParseIP("10.0.8.5"))

	// Process a DISCOVER request
	dp := server.ServeDHCP(p, dhcp4.Discover, p.ParseOptions())
	if dp == nil {
		t.Fatal("Processed packet is nil")
	}

	checkIP(dp, []byte{0xa, 0x0, 0x8, 0x78}, t)
}
