# Global Section

The global section contains settings that will be applied to every network, subnet, pool, and host unless overridden elsewhere. This is where the global, default DNS servers, domain name, and other DHCP options would be specified. The global section is also the only valid place for the `server-identifier` statement.

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

Registered and unregistered blocks may be specified in the global section and like elsewhere will only be applied to their respective lease types. All options/settings are valid here except `server-identifier`.
