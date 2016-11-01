# Network Section

A network is a logical grouping of different subnets. It doesn't have any influence on how leases are given out. It's simply for organizational purposes. In particular when used with Packet Guardian. Although all subnets could be declared inside a single network block, it's advised to separate them logically for readability, performance, and to better see true network usage.

## Network Syntax

The network block has a few different syntax forms. Which one to use entirely depends on readability and circumstances.

Full Form:

```
network NetworkName
    unregistered
        [subnet blocks]
    end
    registered
        [subnet blocks]
    end
end
```

Short Form:

```
network NetworkName
    [subnet blocks]
end
```

Hybrid Form:

```
network NetworkName
    [subnet blocks]
    registered
        [subnet blocks]
    end
end
```

The start line syntax is `network [name]`. The name is completely arbitrary but must be unique to each network block. Names are case insensitive and are lowercased when parsed. Options may be specified within a network block in which case they will apply to both registered and unregistered subnets in that network. A network may contain one or more registered/unregistered blocks and one or more subnets. Subnets outside of a registered/unregistered block are assumed to be unregistered. Although multiple registered/unregistered blocks may be declared, it's considered best practice to have only one of each. All options from multiple blocks will be consolidated meaning if two registered blocks are created each with different options, those options combined will apply to all registered subnets.


## Pools

A pool splits a subnet into multiple ranges from which leases will be given out. Pool blocks may contain any valid options/settings. Each pool must contain only one range statement with the syntax `range [start address] [end address]`. The range is inclusive. See the `Subnets` section for pool block syntax.

## Subnets

A subnet block forms the fundamental building block for the server. Each subnet must be inside a network block. If it's not within a registered/unregistered block, it's assumed to be unregistered. A subnet may contain any valid options. A subnet must have at least one pool, but can have more if desired.

## Subnet Syntax

Like the network block, there's a few different syntax forms for a subnet block. Every subnet block begins with the keyword `subnet` followed by the subnet range in CIDR notation. The simplest is a single pool within a subnet:

```
subnet 10.0.1.0/24
    range 10.0.1.10 10.0.1.200
    option router 10.0.1.1
end
```

If more than one pool is needed, the syntax can get a bit more exotic. Here's the standard, full syntax for multiple pools:

```
subnet 10.0.1.0/24
    option router 10.0.1.1
    pool
        range 10.0.1.10 10.0.1.100
    end
    pool
        range 10.0.1.150 10.0.1.200
        option domain-name example.com
    end
end
```

In the above example, the router option will be given to every lease in both pools. However, only leases from the second pool will be given the domain-name option of "example.com". The first pool will either use a domain-name that was specified earlier, or won't send one at all.

If all the pools will have the same settings/options, a shorter syntax can be used:

```
subnet 10.0.1.0/24
    option router 10.0.1.1
    range 10.0.1.10 10.0.1.100
    range 10.0.1.150 10.0.1.200
end
```

If the above example, leases from both pools will be given the router option.

**WARNING**: When using this shorter, multi-pool syntax, make sure the range statements are at the end of the subnet. Otherwise, any options/settings will be placed inside the pool itself and not the subnet. For example:

```
subnet 10.0.1.0/24
    range 10.0.1.10 10.0.1.100
    option router 10.0.1.1
    range 10.0.1.150 10.0.1.200
end
```

would be equivalent to this full syntax form:

```
subnet 10.0.1.0/24
    pool
        range 10.0.1.10 10.0.1.100
        option router 10.0.1.1
    end
    pool
        range 10.0.1.150 10.0.1.200
    end
end
```

Notice how the option is actually placed inside the first pool and not the subnet "root". This form is perfectly valid, but can be tricky if one is not prepared. If using this form intentionally, it's recommend to put a blank line between the pools and use comments to explain what the pool is for.

The `pool` keyword cannot be used in the same subnet where any short form is also used. The following example is invalid:

```
subnet 10.0.1.0/24
    range 10.0.1.10 10.0.1.30
    option router 10.0.1.1
    pool
        range 10.0.1.40 10.0.1.100
    end
    pool
        range 10.0.1.150 10.0.1.200
    end
end
```

To fix this, either remove the pool/end keywords, or surround the first range with pool/end:

```
subnet 10.0.1.0/24
    range 10.0.1.10 10.0.1.30
    option router 10.0.1.1
    range 10.0.1.40 10.0.1.100
    range 10.0.1.150 10.0.1.200
end

# OR

subnet 10.0.1.0/24
    pool
        range 10.0.1.10 10.0.1.30
        option router 10.0.1.1
    end
    pool
        range 10.0.1.40 10.0.1.100
    end
    pool
        range 10.0.1.150 10.0.1.200
    end
end
```
