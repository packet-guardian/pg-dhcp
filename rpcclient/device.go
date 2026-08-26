package rpcclient

import (
	"net"

	"github.com/packet-guardian/pg-dhcp/models"
)

type DeviceRPCRequest struct {
	client *RPCClient
}

func (d *DeviceRPCRequest) Get(mac net.HardwareAddr) (*models.Device, error) {
	reply := new(models.Device)
	if err := d.client.call("Device.Get", mac, reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *DeviceRPCRequest) Register(mac net.HardwareAddr) error {
	return d.client.call("Device.Register", mac, nil)
}

func (d *DeviceRPCRequest) Unregister(mac net.HardwareAddr) error {
	return d.client.call("Device.Unregister", mac, nil)
}

func (d *DeviceRPCRequest) Blacklist(mac net.HardwareAddr) error {
	return d.client.call("Device.Blacklist", mac, nil)
}

func (d *DeviceRPCRequest) RemoveBlacklist(mac net.HardwareAddr) error {
	return d.client.call("Device.RemoveBlacklist", mac, nil)
}

func (d *DeviceRPCRequest) Delete(mac net.HardwareAddr) error {
	return d.client.call("Device.Delete", mac, nil)
}
