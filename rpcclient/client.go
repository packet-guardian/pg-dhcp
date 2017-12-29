package rpcclient

import (
	"net/rpc"
)

// Client is an RPC client connection to the managed DHCP server
type RPCClient struct {
	c *rpc.Client
}

// Connect creates a new RPC client connection to specified network address.
func Connect(network, address string) (Client, error) {
	c, err := rpc.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &RPCClient{c: c}, nil
}

// Close the RPC connection
func (c *RPCClient) Close() error {
	return c.c.Close()
}

// Device creates a new request to the Device service.
func (c *RPCClient) Device() DeviceRequest {
	return &DeviceRPCRequest{client: c}
}

// Lease creates a new request to the Lease service.
func (c *RPCClient) Lease() LeaseRequest {
	return &LeaseRPCRequest{client: c}
}

// Network creates a new request to the Network service.
func (c *RPCClient) Network() NetworkRequest {
	return &NetworkRPCRequest{client: c}
}

// Server creates a new request to the Server service.
func (c *RPCClient) Server() ServerRequest {
	return &ServerRPCRequest{client: c}
}
