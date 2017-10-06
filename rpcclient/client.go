package rpcclient

import (
	"net/rpc"
)

// Client is an RPC client connection to the managed DHCP server
type Client struct {
	c *rpc.Client
}

// Connect creates a new RPC client connection to specified network address.
func Connect(network, address string) (*Client, error) {
	c, err := rpc.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &Client{c: c}, nil
}

// Close the RPC connection
func (c *Client) Close() error {
	return c.c.Close()
}

// Device creates a new request to the Device service.
func (c *Client) Device() *DeviceRequest {
	return &DeviceRequest{client: c}
}

// Lease creates a new request to the Lease service.
func (c *Client) Lease() *LeaseRequest {
	return &LeaseRequest{client: c}
}

// Network creates a new request to the Network service.
func (c *Client) Network() *NetworkRequest {
	return &NetworkRequest{client: c}
}

// Server creates a new request to the Server service.
func (c *Client) Server() *ServerRequest {
	return &ServerRequest{client: c}
}
