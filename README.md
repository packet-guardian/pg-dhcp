# PG-DHCP

This is the DHCP server package backing the Packet Guardian captive portal. It has been separated into it's own repository to make development a bit easier, and to provide a better focus to the origin project. This package may be used completely independent of Packet Guardian.

Features:

- RFC2131 DHCP protocol
- The most used options are implement, more to come
- Seperation of registered vs unregistered devices (known/unknown)
- Storage independent (the calling project is responsible for storage)
