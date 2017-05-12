// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"bytes"
	"testing"
	"time"

	"github.com/lfkeitel/verbose"
	"github.com/packet-guardian/pg-dhcp/events"
	"github.com/packet-guardian/pg-dhcp/verification"
)

func TestIPGiveOut(t *testing.T) {
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

	pool := c.networks["network1"].subnets[0].pools[0]
	lease := pool.getFreeLease(sc)
	if !bytes.Equal(lease.IP.To4(), []byte{0xa, 0x0, 0x1, 0xa}) {
		t.Errorf("Incorrect lease. Expected %v, got %v", []byte{0xa, 0x0, 0x2, 0xa}, lease.IP)
	}
	lease.End = time.Now().Add(time.Duration(10) * time.Second)

	// Test next lease is given
	lease = pool.getFreeLease(sc)
	if !bytes.Equal(lease.IP.To4(), []byte{0xa, 0x0, 0x1, 0xb}) {
		t.Errorf("Incorrect lease. Expected %v, got %v", []byte{0xa, 0x0, 0x2, 0xb}, lease.IP)
	}
}

func BenchmarkLeaseGiveOutLastLeaseNet24(b *testing.B) {
	benchmarkPool("network1", b)
}

func BenchmarkLeaseGiveOutLastLeaseNet22(b *testing.B) {
	benchmarkPool("network2", b)
}

func benchmarkPool(name string, b *testing.B) {
	db, err := setUpLeaseStore()
	if err != nil {
		b.Fatal(err)
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
		b.Fatalf("Test config failed parsing: %v", err)
	}

	pool := c.networks[name].subnets[0].pools[0]
	// Burn through all but the last lease
	for i := 0; i < pool.getCountOfIPs()-1; i++ {
		lease := pool.getFreeLease(sc)
		if lease == nil {
			b.FailNow()
		}
		lease.End = time.Now().Add(time.Duration(100) * time.Second)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if l := pool.getFreeLease(sc); l == nil {
			b.Fatal("Lease is nil")
		}
	}
}
