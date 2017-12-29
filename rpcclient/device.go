package rpcclient

import (
	"net"

	"github.com/packet-guardian/pg-dhcp/store"
)

type DeviceRPCRequest struct {
	client *RPCClient
}

func (d *DeviceRPCRequest) Get(mac net.HardwareAddr) (*store.Device, error) {
	reply := new(store.Device)
	if err := d.client.c.Call("Device.Get", mac, reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *DeviceRPCRequest) Register(mac net.HardwareAddr) error {
	return d.client.c.Call("Device.Register", mac, nil)
}

func (d *DeviceRPCRequest) Unregister(mac net.HardwareAddr) error {
	return d.client.c.Call("Device.Unregister", mac, nil)
}

func (d *DeviceRPCRequest) Blacklist(mac net.HardwareAddr) error {
	return d.client.c.Call("Device.Blacklist", mac, nil)
}

func (d *DeviceRPCRequest) RemoveBlacklist(mac net.HardwareAddr) error {
	return d.client.c.Call("Device.RemoveBlacklist", mac, nil)
}

func (d *DeviceRPCRequest) Delete(mac net.HardwareAddr) error {
	return d.client.c.Call("Device.Delete", mac, nil)
}
