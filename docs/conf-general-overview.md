# Network Configuration Overview

The DHCP configuration file syntax is a custom syntax loosely based the DHCPD format. The sample DHCP configuration includes explanations and examples of the possible formats. The DHCP server is customized for a registration system where devices are separated based on registration status. However, this application can run a standard DHCP server as well. It adheres to RFC 2131/2132. It currently does not implement any options from other RFCs. Those will come with time.

Options which allow for multiple values such as domain-name-server and network-time-protocol-servers, must be a list of values separated by a space. E.g: `option domain-name-server 10.1.0.1 10.1.0.2`.

For boolean/toggle options, valid values are `true` or `false`.

## Options

Options start with the keyword `option` followed by the option name and finally its value(s). The available options are:

- `subnet-mask`
- `time-offset`
- `router`
- `time-server`
- `name-server`
- `domain-name-server`
- `log-server`
- `cookie-server`
- `lpr-server`
- `impress-server`
- `resource-location-server`
- `hostname`
- `boot-file-size`
- `merit-dump-file`
- `domain-name`
- `swap-server`
- `root-path`
- `extensions-path`
- `ip-forwarding-toggle`
- `non-local-source-routing-toggle`
- `policy-filter`
- `max-datagram-reassembly-size`
- `default-ip-ttl`
- `path-mtu-aging-timeout`
- `path-mtu-plateau-table`
- `interface-mtu`
- `all-subnets-are-local`
- `broadcast-address`
- `perform-mask-discovery`
- `mask-supplier`
- `perform-router-discovery`
- `router-solicitation-address`
- `static-route`
- `trailer-encapsulation`
- `arp-cache-timeout`
- `ethernet-encapsulation`
- `tcp-default-ttl`
- `tcp-keepalive-interval`
- `tcp-keepalive-garbage`
- `network-information-service-domain`
- `network-information-servers`
- `network-time-protocol-servers`
- `vendor-specific-information`
- `netbios-over-tcpip-name-server`
- `netbios-over-tcpip-datagram-distribution-server`
- `netbios-over-tcpip-node-type`
- `netbios-over-tcpip-scope`
- `xwindow-system-font-server`
- `xwindow-system-display-manager`
- `nis+-Domain`
- `nis+-Servers`
- `mobile-ip-home-agent`
- `simple-mail-transport-protocol`
- `post-office-protocol-server`
- `network-news-transport-protocol`
- `default-www-server`
- `default-finger-server`
- `default-irc-server`
- `street-talk-server`
- `street-talk-directory-assistance`
- `tftp-server-name`
- `renewal-time-value`
- `rebinding-time-value`

The following options do NOT begin with the `option` keyword:

- `default-lease-time` - The amount of time in seconds a lease will be active for. Defaults to 12 hours.
- `max-lease-time` - The maximum amount of time in seconds a lease will be active for. Defaults to 12 hours.
- `free-lease-after` - The time in seconds that a lease will be paired with a client MAC address. If a client requests an address after this time, it is not guaranteed they will be given the same lease. This option will only take affect when declared inside a registered and/or unregistered block within the global block.
