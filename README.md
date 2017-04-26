# PG-DHCP

[![GoDoc](https://godoc.org/github.com/packet-guardian/pg-dhcp?status.svg)](https://godoc.org/github.com/packet-guardian/pg-dhcp)
[![GitHub issues](https://img.shields.io/github/issues/packet-guardian/pg-dhcp.svg)](https://github.com/packet-guardian/pg-dhcp/issues)
[![GitHub stars](https://img.shields.io/github/stars/packet-guardian/pg-dhcp.svg)](https://github.com/packet-guardian/pg-dhcp/stargazers)
[![GitHub license](https://img.shields.io/badge/license-New%20BSD-blue.svg)](https://raw.githubusercontent.com/packet-guardian/pg-dhcp/master/LICENSE)

This is the DHCP server package backing the Packet Guardian captive portal. It has been separated into it's own repository to make development a bit easier, and to provide a better focus to the origin project. This package may be used completely independently of Packet Guardian.

Features:

- RFC2131 DHCP protocol
- The most used options are implement, more to come
- Separation of registered vs unregistered devices (known/unknown)
- Storage independent (the calling project is responsible for storage)

[Configuration File Format](configurationFileFormat.md)
