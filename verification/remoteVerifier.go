package verification

import (
	"errors"
	"net"
	"strings"
	"time"

	"github.com/lfkeitel/verbose"
)

const connRWTimeout = 5 * time.Second

type RemoteVerifier struct {
	c              net.Conn
	proto, address string
	logger         *verbose.Logger
	reconnectTO    time.Duration
}

func NewRemoteVerifier(address string, logger *verbose.Logger, reconnectTimeout time.Duration) (*RemoteVerifier, error) {
	addressParts := strings.Split(address, "://")
	if len(addressParts) != 2 {
		return nil, errors.New("Invalid verification address")
	}

	if addressParts[0] != "tcp" && addressParts[0] != "unix" {
		return nil, errors.New("Only TCP and Unix are allowed verification address types")
	}

	r := &RemoteVerifier{
		proto:       addressParts[0],
		address:     addressParts[1],
		logger:      logger,
		reconnectTO: reconnectTimeout,
	}
	if err := r.dial(); err != nil {
		return nil, err
	}
	return r, nil
}

func (v *RemoteVerifier) dial() error {
	conn, err := net.DialTimeout(v.proto, v.address, 10*time.Second)
	if err != nil {
		return err
	}
	v.c = conn
	return nil
}

func (v *RemoteVerifier) redial() error {
	v.logger.Error("Connection to verification server closed, reconnecting...")
	timeout := time.NewTimer(v.reconnectTO)

	for {
		if err := v.dial(); err == nil {
			return nil
		}

		select {
		case <-timeout.C:
			v.logger.Error("Giving up reconnecting to verification server")
			return errors.New("error communicating with verification server")
		case <-time.After(2 * time.Second):
		}

		v.logger.Error("Failed connecting to verification server. Trying again...")
	}
}

func (v *RemoteVerifier) VerifyClient(mac net.HardwareAddr) (ClientStatus, error) {
	v.c.SetDeadline(time.Now().Add(connRWTimeout))

	// Version 1
	// Packet type V
	// Mac address
	_, err := v.c.Write([]byte{1, 'V', mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]})
	if err != nil {
		v.logger.WithField("err", err).Error("failed to write to connection")
		v.c.Close()
		if err := v.redial(); err != nil {
			return ClientDrop, err
		}
		return v.VerifyClient(mac)
	}

	// Version 1
	// Packet type R
	// Response 1 byte
	buf := make([]byte, 3)
	n, err := v.c.Read(buf)
	if err != nil {
		v.logger.WithField("err", err).Error("failed to read from connection")
		v.c.Close()
		if err := v.redial(); err != nil {
			return ClientDrop, err
		}
		return v.VerifyClient(mac)
	}
	if n != 3 || buf[0] != 1 || buf[1] != 'R' {
		return ClientUnregistered, errors.New("Invalid verification response")
	}

	var resp ClientStatus

	switch buf[2] {
	case 'R':
		resp = ClientRegistered
	case 'U':
		resp = ClientUnregistered
	case 'D':
		resp = ClientDrop
	}

	return resp, nil
}
