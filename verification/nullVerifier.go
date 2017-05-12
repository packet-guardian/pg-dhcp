package verification

import "net"

type NullVerifier struct{}

func NewNullVerifier() *NullVerifier {
	return &NullVerifier{}
}

func (v *NullVerifier) VerifyClient(mac net.HardwareAddr) (ClientStatus, error) {
	return ClientUnregistered, nil
}
