# Application Configuration

The application uses TOML for its configuration file.

## Available Settings

```toml
[logging]
Disabled = true         # Disable logging to a file
Level    = "debug"      # Logging level: debug, info, notice, warning, error, critical, alert, emergency, fatal
Path     = "dhcp.log"   # Path of log file

[Database]
Type     = "boltdb"      # Storage type "boltdb", "memory", "mysql", "pg"
Path     = "database.db" # Path to database for "boltdb"
Username = "root"        # Username for "mysql" and "pg"
Password = "password"    # Password for "mysql" and "pg"
Protocol = "tcp"         # Protocol for "mysql" and "pg"
Address  = "localhost"   # Address for "mysql" and "pg"
Port     = 3306          # Port for "mysql" and "pg"
Name     = "pg"          # Database name for "mysql" and "pg"

# Unless you have a VERY good reason, don't change these.
LeaseTable     = "lease"        # Lease table for "pg"
DeviceTable    = "device"       # Device table for "pg"
BlacklistTable = "blacklist"    # Blacklist table for "pg"

[leases]
DeleteAfter = "96h"     # Duration after which old leases are deleted, Go's time.Duration syntax

[server]
BlockBlacklisted = false            # Completely block blacklisted devices
NetworksFile     = "networks.conf"  # Path to network definition file
Workers          = 4                # Number of request workers

[management]
Address    = 0.0.0.0        # IP address to expose management API
Port       = 8677           # Port to expose management API
AllowedIPs = ["10.2.3.5"]   # List of IP addresses that can access the management API
```

## Storage Options

### BoltDB

BoltDB is a file-based, blob storage that's typically much more performant than SQLite or other
file-based databases for blob storage. It's a pure Go implementation which means it's very efficient
and can easily be used cross-platform. This is a good choice for small installations or when running
as a stand-alone server.

### MySQL / MariaDB

Using a typical SQL database can make manual management much easier. BoltDB doesn't currently have a
canonical tool to manage Bolt databases. Also, only one process can use a Bolt database at a time.
Keep in mind that the server uses heavy internal caching especially for leases. Any manual changes to
leases while the server is running will be overwritten or at best completely ignored until the server
is restarted.

**Note**: The MySQL server must run in ANSI mode. This can achieved by running mysql with the `--ansi`
flag to editing the configuration file and adding `sql-mode = "ANSI"` to the `[mysqld]` section.

### PG (Packet Guardian)

This storage type is a modified version of the MySQL storage which allows using an existing Packet Guardian
database instead of having to migrate to a separate database on upgrade.

When using this storage, the management API is downgraded to limited, read-only functionality. Any calls to
alter a Device object will succeed but not do anything. This is because Devices are managed by the Packet
Guardian registration system and not the DHCP server.
