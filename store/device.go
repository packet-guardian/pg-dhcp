package store

import (
	"net"
)

type Device struct {
	MAC         net.HardwareAddr
	Registered  bool
	Blacklisted bool
}
