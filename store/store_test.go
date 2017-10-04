package store

import (
	"bytes"
	"net"
	"os"
	"reflect"
	"testing"
)

func setUpStore() (*Store, error) {
	return NewStore("test.db")
}

func tearDownStore(db *Store) {
	db.Close()
	os.Remove("test.db")
}

func TestLeaseStore(t *testing.T) {
	store, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(store)

	lease := leaseTests[0].actual
	if err := store.PutLease(lease); err != nil {
		t.Fatal(err)
	}
	store.Flush()

	lease2, err := store.GetLease(lease.IP)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(lease, lease2) {
		t.Fatalf("Leases don't match")
	}
}

func TestForEachLease(t *testing.T) {
	store, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(store)

	lease1 := leaseTests[0].actual
	lease2 := leaseTests[1].actual

	store.PutLease(lease1)
	store.PutLease(lease2)
	store.Flush()

	var newLease1, newLease2 *Lease

	store.ForEachLease(func(l *Lease) {
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

func TestDeviceStore(t *testing.T) {
	store, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(store)

	device := &Device{
		MAC:         net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef}),
		Registered:  true,
		Blacklisted: false,
	}
	store.PutDevice(device)
	store.Flush()

	device2 := store.GetDevice(device.MAC)

	if !reflect.DeepEqual(device, device2) {
		t.Fatalf("Devices don't match")
	}
}

func TestDeviceStoreNonExistantDevice(t *testing.T) {
	store, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(store)

	mac := net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})
	device := store.GetDevice(mac)

	if device.Registered {
		t.Fatal("Non existant device shouldn't be registered")
	}
	if device.Blacklisted {
		t.Fatal("Non existant device shouldn't be blacklisted")
	}
}

func TestForEachDevice(t *testing.T) {
	store, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(store)

	device1 := &Device{
		MAC:         net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef}),
		Registered:  true,
		Blacklisted: false,
	}
	device2 := &Device{
		MAC:         net.HardwareAddr([]byte{0x22, 0x34, 0x56, 0xab, 0xcd, 0xef}),
		Registered:  false,
		Blacklisted: true,
	}

	store.PutDevice(device1)
	store.PutDevice(device2)
	store.Flush()

	var newDevice1, newDevice2 *Device

	store.ForEachDevice(func(d *Device) {
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
