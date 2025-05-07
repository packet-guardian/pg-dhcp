// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sconfig

import (
	"bytes"
	"testing"
	"time"
)

func TestIPGiveOut(t *testing.T) {
	db, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(db)

	// Setup Configuration
	c, err := ParseFile("../testdata/testConfig.conf")
	if err != nil {
		t.Fatalf("Test config failed parsing: %v", err)
	}

	pool := c.Networks["network1"].Subnets[0].Pools[0]
	lease := pool.GetFreeLease()
	if !bytes.Equal(lease.IP.To4(), []byte{0xa, 0x0, 0x1, 0xa}) {
		t.Errorf("Incorrect lease. Expected %v, got %v", []byte{0xa, 0x0, 0x2, 0xa}, lease.IP)
	}
	lease.End = time.Now().Add(time.Duration(10) * time.Second)

	// Test next lease is given
	lease = pool.GetFreeLease()
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
	db, err := setUpStore()
	if err != nil {
		b.Fatal(err)
	}
	defer tearDownStore(db)

	// Setup Configuration
	c, err := ParseFile("./testdata/testConfig.conf")
	if err != nil {
		b.Fatalf("Test config failed parsing: %v", err)
	}

	pool := c.Networks[name].Subnets[0].Pools[0]
	// Burn through all but the last lease
	for i := 0; i < pool.GetCountOfIPs()-1; i++ {
		lease := pool.GetFreeLease()
		if lease == nil {
			b.FailNow()
		}
		lease.End = time.Now().Add(time.Duration(100) * time.Second)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if l := pool.GetFreeLease(); l == nil {
			b.Fatal("Lease is nil")
		}
	}
}
