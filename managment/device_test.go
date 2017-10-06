package management

import (
	"net"
	"testing"

	"github.com/packet-guardian/pg-dhcp/store"
)

func TestGetDeviceRPC(t *testing.T) {
	db, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(db)

	mac := net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})
	db.PutDevice(&store.Device{
		MAC:        mac,
		Registered: true,
	})

	device := &Device{store: db}
	d := new(store.Device)

	device.Get(mac, d)
	if !d.Registered {
		t.Fatal("Device wasn't getted.")
	}
}

func TestRegisterDeviceRPC(t *testing.T) {
	db, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(db)

	mac := net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})

	device := &Device{store: db}
	var ack bool
	device.Register(mac, &ack)

	d := db.GetDevice(mac)
	if !d.Registered {
		t.Fatal("Device wasn't registered.")
	}
}

func TestUnregisterUnblacklistedDeviceRPC(t *testing.T) {
	db, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(db)

	mac := net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})

	db.PutDevice(&store.Device{
		MAC:        mac,
		Registered: true,
	})

	device := &Device{store: db}

	d := db.GetDevice(mac)
	if !d.Registered {
		t.Fatal("Device wasn't stored properly.")
	}

	var ack bool
	device.Unregister(mac, &ack)
	d = db.GetDevice(mac)
	if d.Registered {
		t.Fatal("Device wasn't unregistered.")
	}
}

func TestUnregisterBlacklistedDeviceRPC(t *testing.T) {
	db, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(db)

	mac := net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})

	db.PutDevice(&store.Device{
		MAC:         mac,
		Registered:  true,
		Blacklisted: true,
	})

	device := &Device{store: db}

	d := db.GetDevice(mac)
	if !d.Registered || !d.Blacklisted {
		t.Fatal("Device wasn't stored properly.")
	}

	var ack bool
	device.Unregister(mac, &ack)
	d = db.GetDevice(mac)
	if d.Registered {
		t.Fatal("Device wasn't unregistered.")
	}
	if !d.Blacklisted {
		t.Fatal("Device was deleted when it shouldn't have been.")
	}
}

func TestBlacklistDeviceRPC(t *testing.T) {
	db, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(db)

	mac := net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})

	device := &Device{store: db}
	var ack bool
	device.Blacklist(mac, &ack)

	d := db.GetDevice(mac)
	if !d.Blacklisted {
		t.Fatal("Device wasn't blacklisted.")
	}
}

func TestRemoveBlacklistRegisteredDeviceRPC(t *testing.T) {
	db, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(db)

	mac := net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})

	db.PutDevice(&store.Device{
		MAC:         mac,
		Registered:  true,
		Blacklisted: true,
	})

	device := &Device{store: db}

	d := db.GetDevice(mac)
	if !d.Registered || !d.Blacklisted {
		t.Fatal("Device wasn't stored properly.")
	}

	var ack bool
	device.RemoveBlacklist(mac, &ack)
	d = db.GetDevice(mac)
	if d.Blacklisted {
		t.Fatal("Device wasn't removed from the blacklist.")
	}
	if !d.Registered {
		t.Fatal("Device was deleted when it shouldn't have been.")
	}
}

func TestDeleteDeviceRPC(t *testing.T) {
	db, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownStore(db)

	mac := net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})

	db.PutDevice(&store.Device{
		MAC:        mac,
		Registered: true,
	})

	device := &Device{store: db}

	d := db.GetDevice(mac)
	if !d.Registered {
		t.Fatal("Device wasn't stored properly.")
	}

	var ack bool
	device.Delete(mac, &ack)
	d = db.GetDevice(mac)
	if d.Registered {
		t.Fatal("Device wasn't deleted.")
	}
}
