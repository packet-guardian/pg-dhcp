package verification

import "net"

type Verifier interface {
	VerifyClient(net.HardwareAddr) (ClientStatus, error)
}

type ClientStatus int

const (
	ClientRegistered ClientStatus = iota
	ClientUnregistered
	ClientDrop
)
