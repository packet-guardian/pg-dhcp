// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package store

import (
	"errors"
	"net"
	"time"

	"github.com/packet-guardian/pg-dhcp/internal/utils"
)

var errBufTooSmall = errors.New("buffer too small")

// A Lease represents a single DHCP lease in a pool. It is bound to a particular
// pool and network.
type Lease struct {
	IP          net.IP
	MAC         net.HardwareAddr
	Network     string
	Start       time.Time
	End         time.Time
	Hostname    string
	IsAbandoned bool
	Offered     bool
	Registered  bool
	Used        bool
}

func NewLease() *Lease {
	return &Lease{}
}

// IsFree determines if the lease is expired and available for use
func (l *Lease) IsFree() bool {
	return (l.Used || l.IsExpired())
}

func (l *Lease) IsExpired() bool {
	return l.End.Before(time.Now())
}

func (l *Lease) Serialize() []byte {
	netBytes := []byte(l.Network)
	hostnameBytes := []byte(l.Hostname)
	buf := make([]byte, 29+len(netBytes)+len(hostnameBytes))

	// IPv4 Address
	copy(buf[:4], l.IP.To4())

	// MAC Address
	copy(buf[4:10], l.MAC)

	// Boolean fields
	if l.IsAbandoned {
		buf[10] = 1
	}
	if l.Registered {
		buf[11] = 1
	}

	// Start time as int64
	copy(buf[12:20], utils.Itob(l.Start.Unix()))
	// End time as int64
	copy(buf[20:28], utils.Itob(l.End.Unix()))

	// Length of network name
	buf[28] = byte(len(netBytes))
	netEnd := len(netBytes) + 29
	// Network name
	copy(buf[29:netEnd], netBytes)

	// Hostname. Hostname as no length as it's everything after the network name
	copy(buf[netEnd:], hostnameBytes)
	return buf
}

func (l *Lease) Unserialize(data []byte) error {
	if len(data) < 29 {
		return errBufTooSmall
	}

	// IP Address
	l.IP = net.IP(make([]byte, 4))
	copy(l.IP, data[:4])

	// MAC Address
	l.MAC = net.HardwareAddr(make([]byte, 6))
	copy(l.MAC, data[4:10])

	// Boolean fields
	l.IsAbandoned = (data[10] == 1)
	l.Registered = (data[11] == 1)

	// Start time as int64
	l.Start = time.Unix(utils.Btoi(data[12:20]), 0)
	// End time as int64
	l.End = time.Unix(utils.Btoi(data[20:28]), 0)

	// Network name
	netlen := int(data[28])
	if len(data) < 29+netlen {
		return errBufTooSmall
	}
	if netlen > 0 {
		l.Network = string(data[29 : netlen+29])
	}

	if len(data) > 29+netlen {
		// Hostname
		hostnameStart := netlen + 29
		l.Hostname = string(data[hostnameStart:])
	}
	return nil
}
