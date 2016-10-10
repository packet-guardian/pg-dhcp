# DHCP Configuration

## Overview

The DHCP configuration file syntax is a custom syntax loosely based the DHCPD format. The sample DHCP configuration includes explanations and examples of the possible formats. The DHCP server is customized for a registration system where devices are separated based on registration status. However, this package can be used to implement a normal DHCP server as well. It adheres to the base DHCP RFC, but not all DHCP options are implemented. Implemented options are described below.

Options which allow for multiple values such as domain-name-server and network-time-protocol-servers, must be a list of values separated by a space. E.g: `option domain-name-server 10.1.0.1 10.1.0.2`.

To specify multiple scopes in a single subnet, each scope must be in its own pool. This is contrary to DHCPD where a single pool can contain multiple range statements. See the Subnet and Pool sections for more.

## Global

Sample:

```
global
    option domain-name example.com
    server-identifier 10.0.0.1

    registered
        free-lease-after 172800
        default-lease-time 86400
        max-lease-time 86400
        option domain-name-server 10.1.0.1 10.1.0.2
    end

    unregistered
        free-lease-after 600
        default-lease-time 360
        max-lease-time 360
        option domain-name-server 10.0.0.1
    end
end
```

The global section contains three subsections. The "root" which is any statement outside a `registered` or `unregistered` block. Options specified here will be applied to every subnet regardless of registration status unless overridden elsewhere. Global root specific statements are:

- `server-identifier` - The IP address of the DHCP server

Registered and unregistered blocks may be located inside a global or network block. Settings here will apply to either registered or unregistered leases either globally (if in global block) or for any containing subnets. These blocks may contain any option along with the lease time settings.

## Options

Most options correspond to a DHCP option and begin with the keyword `option`. The available options are:

- `subnet-mask`
- `router`
- `domain-name-server` - Multiple IP addresses must be separated with a space
- `domain-name`
- `broadcast-address`
- `network-time-protocol-servers` - Multiple IP addresses must be separated with a space

The following options do NOT begin with the `option` keyword:

- `default-lease-time` - The amount of time in seconds a lease will be active for. Defaults to 12 hours.
- `max-lease-time` - The maximum amount of time in seconds a lease will be active for. Defaults to 12 hours.
- `free-lease-after` - The time in seconds that a lease will be paired with a client MAC address. If a client requests an address after this time, it is not guaranteed they will be given the same lease. This option will only take affect when declared inside a registered and/or unregistered block within the global block.

## Network

```
network Network1
    unregistered
        ...
    end
    registered
        ...
    end
end
```

A network block groups multiple subnets into logical units. Although technically all subnets could be located in a single network block, it would be incredibly inefficient and difficult to determine true network usage.

The start line syntax is `network [name]`. The name is completely arbitrary but must be unique to each network block. Names are case insensitive. Options may be specified within a network block in which case they will apply to both registered and unregistered leases in that network. A network may contain one or more registered/unregistered blocks and one or more subnets. Subnets outside of a registered/unregistered block are assumed to be unregistered. Although multiple registered/unregistered blocks may be declared, it's considered best practice to have only one of each. All options from multiple blocks will be consolidated meaning if two registered blocks are created each with different options, those options combined will apply to all registered subnets.

## Subnet

```
# Shortened pool syntax - One pool/range
subnet 10.0.1.0/24
    range 10.0.1.10 10.0.1.200
    option router 10.0.1.1
end

# Full pool syntax - Multiple pools/ranges
subnet 10.0.2.0/24
    option router 10.0.2.1
    pool
        range 10.0.2.10 10.0.2.100
    end
    pool
        range 10.0.2.150 10.0.2.200
    end
end

# Invalid full pool syntax - range cannot appear outside of a pool block
subnet 10.0.2.0/24
    range 10.0.2.10 10.0.2.30
    option router 10.0.2.1
    pool
        range 10.0.2.40 10.0.2.100
    end
    pool
        range 10.0.2.150 10.0.2.200
    end
end
```

A subnet block forms the fundamental building block for the server. Each subnet must be inside a network block. If it's not within a registered/unregistered block, it's assumed to be unregistered. The start line syntax is `subnet [ip range in CIDR notation]`. A subnet may contain any valid options as described above. A subnet may contain a single pool in which case a single range statement may be given. If multiple pool ranges are needed, the full syntax must be used.

## Pool

A pool splits a subnet into multiple ranges. Typically, the shortened syntax will suffice where only one pool is present in a subnet. If you want multiple address ranges, you will need multiple pool blocks. Pool blocks may contain any valid option as specified above. Each pool must contain one range statement with the syntax `range [start address] [end address]`. The range is inclusive.
