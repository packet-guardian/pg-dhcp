package rpcclient

import (
	"net"

	"github.com/packet-guardian/pg-dhcp/store"
)

type DeviceRequest struct {
	client *Client
}

func (d *DeviceRequest) Get(mac net.HardwareAddr) (*store.Device, error) {
	reply := new(store.Device)
	if err := d.client.c.Call("Device.Get", mac, reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *DeviceRequest) Register(mac net.HardwareAddr) error {
	return d.client.c.Call("Device.Register", mac, nil)
}

func (d *DeviceRequest) Unregister(mac net.HardwareAddr) error {
	return d.client.c.Call("Device.Unregister", mac, nil)
}

func (d *DeviceRequest) Blacklist(mac net.HardwareAddr) error {
	return d.client.c.Call("Device.Blacklist", mac, nil)
}

func (d *DeviceRequest) RemoveBlacklist(mac net.HardwareAddr) error {
	return d.client.c.Call("Device.RemoveBlacklist", mac, nil)
}

func (d *DeviceRequest) Delete(mac net.HardwareAddr) error {
	return d.client.c.Call("Device.Delete", mac, nil)
}
