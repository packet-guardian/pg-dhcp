package management

import (
	"net"

	"github.com/packet-guardian/pg-dhcp/models"
	"github.com/packet-guardian/pg-dhcp/store"
)

type Device struct {
	store store.Store
}

func (d *Device) Get(mac net.HardwareAddr, reply *models.Device) error {
	device, _ := d.store.GetDevice(mac)
	*reply = *device
	return nil
}

func (d *Device) Register(mac net.HardwareAddr, ack *bool) error {
	device, _ := d.store.GetDevice(mac)
	if !device.Registered {
		device.Registered = true
		d.store.PutDevice(device)
	}

	*ack = true
	return nil
}

func (d *Device) Unregister(mac net.HardwareAddr, ack *bool) error {
	device, _ := d.store.GetDevice(mac)
	if device.Registered {
		if device.Blacklisted {
			device.Registered = false
			d.store.PutDevice(device)
		} else {
			d.store.DeleteDevice(device)
		}
	}

	*ack = true
	return nil
}

func (d *Device) Blacklist(mac net.HardwareAddr, ack *bool) error {
	device, _ := d.store.GetDevice(mac)
	if !device.Blacklisted {
		device.Blacklisted = true
		d.store.PutDevice(device)
	}

	*ack = true
	return nil
}

func (d *Device) RemoveBlacklist(mac net.HardwareAddr, ack *bool) error {
	device, _ := d.store.GetDevice(mac)
	if device.Blacklisted {
		if device.Registered {
			device.Blacklisted = false
			d.store.PutDevice(device)
		} else {
			d.store.DeleteDevice(device)
		}
	}

	*ack = true
	return nil
}

func (d *Device) Delete(mac net.HardwareAddr, ack *bool) error {
	device, _ := d.store.GetDevice(mac)
	d.store.DeleteDevice(device)

	*ack = true
	return nil
}
