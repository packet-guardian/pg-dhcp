package store

import (
	"bytes"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/packet-guardian/pg-dhcp/models"
)

var leaseTests = []struct {
	data   []byte
	actual *models.Lease
}{
	{
		data: []byte{0xa, 0x0, 0x2, 0x5, 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x0, 0x1, 0xd0, 0xf9, 0x87, 0x90, 0xb, 0x0, 0x0, 0x0, 0xa0, 0x89, 0x88, 0x90, 0xb, 0x0, 0x0, 0x0, 0xa, 0x4e, 0x65, 0x74, 0x77, 0xc3, 0xb6, 0x72, 0x6b, 0x20, 0x31, 0x53, 0x6f, 0x6d, 0x65, 0x20, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x20, 0x74, 0x68, 0x61, 0x74, 0x20, 0x69, 0x73, 0x20, 0x61, 0x20, 0x64, 0x65, 0x63, 0x65, 0x6e, 0x74, 0x20, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68},
		actual: &models.Lease{
			IP:          net.ParseIP("10.0.2.5").To4(),
			MAC:         net.HardwareAddr([]byte{0xab, 0xcd, 0xef, 0x12, 0x34, 0x56}),
			Network:     "Netwörk 1",
			Start:       time.Unix(1493237352, 0),
			End:         time.Unix(1493238352, 0),
			Hostname:    "Some hostname that is a decent length",
			IsAbandoned: false,
			Registered:  true,
		},
	},
	{
		data: []byte{0xa, 0x0, 0x2, 0x6, 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x0, 0x1, 0xd0, 0xf9, 0x87, 0x90, 0xb, 0x0, 0x0, 0x0, 0xa0, 0x89, 0x88, 0x90, 0xb, 0x0, 0x0, 0x0, 0xa, 0x4e, 0x65, 0x74, 0x77, 0xc3, 0xb6, 0x72, 0x6b, 0x20, 0x31},
		actual: &models.Lease{
			IP:          net.ParseIP("10.0.2.6").To4(),
			MAC:         net.HardwareAddr([]byte{0xab, 0xcd, 0xef, 0x12, 0x34, 0x56}),
			Network:     "Netwörk 1",
			Start:       time.Unix(1493237352, 0),
			End:         time.Unix(1493238352, 0),
			Hostname:    "",
			IsAbandoned: false,
			Registered:  true,
		},
	},
	{
		data: []byte{0xa, 0x0, 0x2, 0x7, 0xab, 0xcd, 0xef, 0x12, 0x34, 0x56, 0x0, 0x1, 0xd0, 0xf9, 0x87, 0x90, 0xb, 0x0, 0x0, 0x0, 0xa0, 0x89, 0x88, 0x90, 0xb, 0x0, 0x0, 0x0, 0x0},
		actual: &models.Lease{
			IP:          net.ParseIP("10.0.2.7").To4(),
			MAC:         net.HardwareAddr([]byte{0xab, 0xcd, 0xef, 0x12, 0x34, 0x56}),
			Network:     "",
			Start:       time.Unix(1493237352, 0),
			End:         time.Unix(1493238352, 0),
			Hostname:    "",
			IsAbandoned: false,
			Registered:  true,
		},
	},
}

func tearDownStore(s Store) error {
	return s.Close()
}

type flusher interface {
	Flush()
}

func testLeaseStore(t *testing.T, s Store) {
	lease := leaseTests[0].actual
	if err := s.PutLease(lease); err != nil {
		t.Fatal(err)
	}
	if f, ok := s.(flusher); ok {
		f.Flush()
	}

	lease2, err := s.GetLease(lease.IP)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(lease, lease2) {
		t.Fatalf("Leases don't match")
	}
}

func testForEachLease(t *testing.T, s Store) {
	lease1 := leaseTests[0].actual
	lease2 := leaseTests[1].actual

	s.PutLease(lease1)
	s.PutLease(lease2)
	if f, ok := s.(flusher); ok {
		f.Flush()
	}

	var newLease1, newLease2 *models.Lease

	s.ForEachLease(func(l *models.Lease) {
		if l.IP.String() == "10.0.2.5" {
			newLease1 = l
		} else if l.IP.String() == "10.0.2.6" {
			newLease2 = l
		}
	})

	if newLease1 == nil {
		t.Error("newLease1 is nil")
	}
	if newLease2 == nil {
		t.Error("newLease2 is nil")
	}
}

func testDeviceStore(t *testing.T, s Store) {
	device := &models.Device{
		MAC:         net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef}),
		Registered:  true,
		Blacklisted: false,
	}
	s.PutDevice(device)

	device2, _ := s.GetDevice(device.MAC)

	if !reflect.DeepEqual(device, device2) {
		t.Fatalf("Devices don't match")
	}
}

func testDeviceStoreNonExistantDevice(t *testing.T, s Store) {
	mac := net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})
	device, _ := s.GetDevice(mac)

	if device.Registered {
		t.Fatal("Non existant device shouldn't be registered")
	}
	if device.Blacklisted {
		t.Fatal("Non existant device shouldn't be blacklisted")
	}
}

func testForEachDevice(t *testing.T, s Store) {
	device1 := &models.Device{
		MAC:         net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef}),
		Registered:  true,
		Blacklisted: false,
	}
	device2 := &models.Device{
		MAC:         net.HardwareAddr([]byte{0x22, 0x34, 0x56, 0xab, 0xcd, 0xef}),
		Registered:  false,
		Blacklisted: true,
	}

	s.PutDevice(device1)
	s.PutDevice(device2)

	var newDevice1, newDevice2 *models.Device

	s.ForEachDevice(func(d *models.Device) {
		if bytes.Equal([]byte(d.MAC), []byte(device1.MAC)) {
			newDevice1 = d
		} else if bytes.Equal([]byte(d.MAC), []byte(device2.MAC)) {
			newDevice2 = d
		}
	})

	if newDevice1 == nil {
		t.Error("newDevice1 is nil")
	}
	if !reflect.DeepEqual(device1, newDevice1) {
		t.Fatalf("device1 and newDevice1 don't match")
	}

	if newDevice2 == nil {
		t.Error("newDevice2 is nil")
	}
	if !reflect.DeepEqual(device2, newDevice2) {
		t.Fatalf("device2 and newDevice2 don't match")
	}
}
