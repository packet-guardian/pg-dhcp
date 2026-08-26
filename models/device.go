package models

import (
	"net"
	"time"
)

type Device struct {
	MAC         net.HardwareAddr
	Registered  bool
	Blacklisted bool
	LastSeen    time.Time
}
