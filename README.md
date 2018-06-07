# PG-DHCP

[![GitHub issues](https://img.shields.io/github/issues/packet-guardian/pg-dhcp.svg)](https://github.com/packet-guardian/pg-dhcp/issues)
[![GitHub stars](https://img.shields.io/github/stars/packet-guardian/pg-dhcp.svg)](https://github.com/packet-guardian/pg-dhcp/stargazers)
[![GitHub license](https://img.shields.io/badge/license-New%20BSD-blue.svg)](https://raw.githubusercontent.com/packet-guardian/pg-dhcp/master/LICENSE)

This is the DHCP server created specifically for registration scenerios. The server exposes a management API that can be
implemented with any other application to manage registration and blocked status of a device. The server can also be
used by itself as a standard DHCP server without any registration information.

## Features

- RFC2131 DHCP protocol
- The most used options are implement, more to come
- Separation of registered vs unregistered devices (known/unknown)
- Multiple storage options including BoldDB and MySQL/MariaDB
- Integrates with version 1 of Packet Guardian

[Documentation](docs)

## Testing

The normal tests do not test MySQL integration. Specifically, the MySQLStore and PGStore are not tested. Integration tests can
be ran using `make integration-test` or by adding the `mysql` tag to `go test`. A MySQL (or compatible) database will need to
be available for use. Connection information can be managed using the following environment variables:

- `MYSQL_TEST_USER` - Default: "root"
- `MYSQL_TEST_PASS` - Default: "password"
- `MYSQL_TEST_PROT` - Default: "tcp"
- `MYSQL_TEST_ADDR` - Default: "localhost:3306"
- `MYSQL_TEST_DBNAME` - Default: "gotest"

There's a docker compose file in `store/testdata` that will spin up a MariaDB container using the defaults above. The test will
create any necessary tables so no additional setup is needed.
