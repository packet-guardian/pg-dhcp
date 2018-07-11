# RPC Management

Starting with version 2.0, the DHCP server exposes a management port that uses RPC
calls for administrative tasks. This page documents the protocol, available
methods, and usage information for the CLI tool.

## Protocol

The RPC connection is handled through Go's RPC library using gob encoding for
transport.

## CLI Usage

Available subcommands:

- `leases`:
    - `-n NETWORK`: Show all leases in a named network
    - `-ip ADDRESS`: Show specific lease information for address
- `networks`: List all network names
- `pools`: Print DHCP pool statistics
- `devices`:
    - `show MAC`: Print information about a specific device
    - `register MAC`: Mark a device as registered
    - `unregister MAC`: Mark a device as unregistered
    - `blacklist MAC`: Mark a device as blacklisted
    - `unblacklist MAC`: Mark a device as not blacklisted
    - `delete MAC`: Delete a device
        - Note: A deleted device will still show information. This is because
        ever MAC addresses creates an implicit, non-persistent device object
        using the application defaults for its various field data.

## RPC Methods

### Lease

- `Lease.Get`
    - **Arguments**: 1 IP address
    - **Result**: Single lease object
    - **Description**: Returns lease information for a specific IP address
- `Lease.GetAllFromNetwork`
    - **Arguments**: 1 string (network name)
    - **Result**: Slice of lease objects
    - **Description**: Returns lease information for all leases in a network

### Network

- `Network.GetNameList`
    - **Arguments**: None
    - **Result**: String slice of network names
    - **Description**: Returns list of network names defined in server

### Server

- `Server.GetPoolStats`
    - **Arguments**: None
    - **Result**: Slice of pool stat objects
    - **Description**: Returns list of pool statistics

### Device

- `Device.Get`
    - **Arguments**: 1 MAC Address
    - **Result**: Single device object
    - **Description**: Returns information about a single device
- `Device.Register`
    - **Arguments**: 1 MAC Address
    - **Result**: None
    - **Description**: Marks device as registered
- `Device.Unregister`
    - **Arguments**: 1 MAC Address
    - **Result**: None
    - **Description**: Marks device as not registered
- `Device.Blacklist`
    - **Arguments**: 1 MAC Address
    - **Result**: None
    - **Description**: Marks device as blacklisted
- `Device.RemoveBlacklist`
    - **Arguments**: 1 MAC Address
    - **Result**: None
    - **Description**: Marks device as not blacklisted
- `Device.Delete`
    - **Arguments**: 1 MAC Address
    - **Result**: None
    - **Description**: Deletes a device
