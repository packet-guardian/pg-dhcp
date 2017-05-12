// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO: Clean up the handler functions. There's a lot of duplicated code that
// could be extracted to a function.

package server

import (
	"bytes"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/lfkeitel/verbose"
	dhcp4 "github.com/onesimus-systems/dhcp4"
	"github.com/packet-guardian/pg-dhcp/store"
	"github.com/packet-guardian/pg-dhcp/verification"
)

var (
	c *Config
)

// A Handler processes all incoming DHCP packets.
type Handler struct {
	gatewayCache map[string]*network
	gatewayMutex sync.Mutex
	c            *ServerConfig
	conn         net.PacketConn
}

// NewDHCPServer creates and sets up a new DHCP Handler with the give configuration.
func NewDHCPServer(conf *Config, s *ServerConfig) *Handler {
	if s.Log == nil {
		s.Log = createLogger()
	}
	c = conf
	return &Handler{
		c:            s,
		gatewayCache: make(map[string]*network),
		gatewayMutex: sync.Mutex{},
	}
}

func createLogger() *verbose.Logger {
	logger := verbose.New("dhcp")

	// Add standard output handler
	sh := verbose.NewStdoutHandler(true)
	logger.AddHandler("stdout", sh)
	sh.SetMinLevel(verbose.LogLevelInfo)

	return logger
}

// ListenAndServe starts the DHCP Handler listening on port 67 for packets.
// This is blocking like HTTP's ListenAndServe method.
func (h *Handler) ListenAndServe() error {
	h.c.Log.Info("Starting DHCP server...")
	l, err := net.ListenPacket("udp4", ":67")
	if err != nil {
		return err
	}
	defer l.Close()
	h.conn = l
	return dhcp4.Serve(l, h)
}

func (h *Handler) Close() error {
	h.conn.Close()
	h.c.Store.Close()
	return nil
}

// LoadLeases will import any current leases saved to the database.
func (h *Handler) LoadLeases() error {
	h.c.Store.ForEachLease(func(l *store.Lease) {
		// Check if the network exists
		n, ok := c.networks[l.Network]
		if !ok {
			return
		}

		// Find the correct pool
		// TODO: Optimize this maybe with a temporary cache
	subnetLoop:
		for _, subnet := range n.subnets {
			if !subnet.includes(l.IP) {
				continue
			}

			for _, pool := range subnet.pools {
				if !pool.includes(l.IP) {
					continue
				}
				pool.leases[l.IP.String()] = l
				h.c.Log.WithField("address", l.IP).Debug("Loaded lease")
				break subnetLoop
			}
		}
	})
	return nil
}

// ServeDHCP processes an incoming DHCP packet and returns a response.
func (h *Handler) ServeDHCP(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) dhcp4.Packet {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			runtime.Stack(buf, false)
			h.c.Log.WithFields(verbose.Fields{
				"package": "dhcp",
				"error":   r,
				"stack":   string(buf),
			}).Critical("Recovering from DHCP panic")
		}
	}()

	// Log every message
	if server, ok := options[dhcp4.OptionServerIdentifier]; !ok || net.IP(server).Equal(c.global.serverIdentifier) {
		h.c.Log.WithFields(verbose.Fields{
			"type":     msgType.String(),
			"ip":       p.CIAddr().String(),
			"mac":      p.CHAddr().String(),
			"relay_ip": p.GIAddr().String(),
		}).Debug()
	}

	var response dhcp4.Packet
	switch msgType {
	case dhcp4.Discover:
		response = h.handleDiscover(p, options)
	case dhcp4.Request:
		response = h.handleRequest(p, options)
	case dhcp4.Release:
		response = h.handleRelease(p, options)
	case dhcp4.Decline:
		response = h.handleDecline(p, options)
	case dhcp4.Inform:
		response = h.handleInform(p, options)
	}
	return response
}

// Handle DHCP DISCOVER messages
func (h *Handler) handleDiscover(p dhcp4.Packet, options dhcp4.Options) dhcp4.Packet {
	// Don't respond to requests on the same subnet
	if p.GIAddr().Equal(net.IPv4zero) {
		return nil
	}

	action, err := h.c.Verification.VerifyClient(p.CHAddr())
	if err != nil {
		h.c.Log.WithField("error", err).Error("Verification error")
		return nil
	}
	if action == verification.ClientDrop {
		return nil
	}
	registered := (action == verification.ClientRegistered)

	// Get network object that the relay IP belongs to
	h.gatewayMutex.Lock()
	network, ok := h.gatewayCache[p.GIAddr().String()]
	if !ok {
		// That gateway hasn't been seen before, find its network
		network = c.searchNetworksFor(p.GIAddr())
		if network == nil {
			h.gatewayMutex.Unlock()
			h.c.Log.WithField("relay_ip", p.GIAddr().String()).Notice("Network not found")
			return nil
		}
		// Add to cache for later
		h.gatewayCache[p.GIAddr().String()] = network
	}
	h.gatewayMutex.Unlock()

	// Find an appropiate lease
	lease, pool := network.getLeaseByMAC(p.CHAddr(), registered)
	if lease == nil {
		// Device doesn't have a recent lease, get a new one
		lease, pool = network.getFreeLease(h.c, registered)
		if lease == nil { // No free lease was found, be more aggressive
			lease, pool = network.getFreeLeaseDesperate(h.c, registered)
		}
		if lease == nil { // Still no lease was found, error and go to the next request
			h.c.Log.WithFields(verbose.Fields{
				"network":    network.name,
				"registered": registered,
			}).Alert("No free leases available in network")
			return nil
		}
	}

	h.c.Log.WithFields(verbose.Fields{
		"ip":         lease.IP.String(),
		"mac":        p.CHAddr().String(),
		"registered": registered,
		"network":    network.name,
		"action":     "offer",
	}).Info("Offering lease to client")

	// Set temporary offered flag and end time
	lease.Offered = true
	lease.Start = time.Now()
	lease.End = time.Now().Add(time.Duration(30) * time.Second) // Set a short end time so it's not offered to other clients
	lease.MAC = make([]byte, len(p.CHAddr()))
	copy(lease.MAC, p.CHAddr())
	// No Save because this is a temporary "lease", if the client accepts then we commit to storage
	// Get options
	leaseOptions := pool.getOptions(registered)
	// Send an offer
	return dhcp4.ReplyPacket(
		p,
		dhcp4.Offer,
		c.global.serverIdentifier,
		lease.IP,
		pool.getLeaseTime(0, registered),
		leaseOptions.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]),
	)
}

// Handle DHCP REQUEST messages
func (h *Handler) handleRequest(p dhcp4.Packet, options dhcp4.Options) dhcp4.Packet {
	if server, ok := options[dhcp4.OptionServerIdentifier]; ok && !net.IP(server).Equal(c.global.serverIdentifier) {
		return nil // Message not for this dhcp server
	}

	reqIP := net.IP(options[dhcp4.OptionRequestedIPAddress])
	if reqIP == nil {
		reqIP = net.IP(p.CIAddr())
	}

	if len(reqIP) != 4 || reqIP.Equal(net.IPv4zero) {
		return dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil)
	}

	action, err := h.c.Verification.VerifyClient(p.CHAddr())
	if err != nil {
		h.c.Log.WithField("error", err).Error("Verification error")
		return nil
	}
	if action == verification.ClientDrop {
		return nil
	}
	registered := (action == verification.ClientRegistered)

	var network *network
	// Get network object that the relay or client IP belongs to
	if p.GIAddr().Equal(net.IPv4zero) {
		// Coming directly from the client
		network = c.searchNetworksFor(reqIP)
	} else {
		// Coming from a relay
		h.gatewayMutex.Lock()
		var ok bool
		network, ok = h.gatewayCache[p.GIAddr().String()]
		h.gatewayMutex.Unlock()
		if !ok {
			// That gateway hasn't been seen before, it needs to go through DISCOVER
			return dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil)
		}
	}

	if network == nil {
		h.c.Log.WithFields(verbose.Fields{
			"ip":         reqIP.String(),
			"registered": registered,
		}).Info("Got a REQUEST for IP not in a scope")
		return dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil)
	}

	lease, pool := network.getLeaseByIP(reqIP, registered)
	if lease == nil || lease.MAC == nil { // If it returns a new lease, the MAC is nil
		h.c.Log.WithFields(verbose.Fields{
			"ip":         reqIP.String(),
			"mac":        p.CHAddr().String(),
			"network":    network.name,
			"registered": registered,
		}).Info("Client tried to request a lease that doesn't exist")
		return dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil)
	}

	if !bytes.Equal(lease.MAC, p.CHAddr()) {
		h.c.Log.WithFields(verbose.Fields{
			"ip":         reqIP.String(),
			"mac":        p.CHAddr().String(),
			"lease_mac":  lease.MAC.String(),
			"network":    network.name,
			"registered": registered,
		}).Info("Client tried to request lease not belonging to them")
		return dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil)
	}

	leaseDur := pool.getLeaseTime(0, registered)
	lease.Start = time.Now()
	lease.End = time.Now().Add(leaseDur + (time.Duration(10) * time.Second)) // Add 10 seconds to account for slight clock drift
	lease.Offered = false
	if ci, ok := options[dhcp4.OptionHostName]; ok {
		lease.Hostname = string(ci)
	}
	if err := h.c.Store.PutLease(lease); err != nil {
		h.c.Log.WithFields(verbose.Fields{
			"mac":   p.CHAddr().String(),
			"error": err,
		}).Error("Error saving lease")
		return dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil)
	}
	leaseOptions := pool.getOptions(registered)

	h.c.Log.WithFields(verbose.Fields{
		"ip":         lease.IP.String(),
		"mac":        lease.MAC.String(),
		"duration":   leaseDur.String(),
		"network":    network.name,
		"registered": registered,
		"hostname":   lease.Hostname,
		"action":     "request_ack",
	}).Info("Acknowledging request")

	return dhcp4.ReplyPacket(
		p,
		dhcp4.ACK,
		c.global.serverIdentifier,
		lease.IP,
		leaseDur,
		leaseOptions.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]),
	)
}

// Handle DHCP RELEASE messages
func (h *Handler) handleRelease(p dhcp4.Packet, options dhcp4.Options) dhcp4.Packet {
	reqIP := p.CIAddr()
	if reqIP == nil || reqIP.Equal(net.IPv4zero) {
		return nil
	}

	action, err := h.c.Verification.VerifyClient(p.CHAddr())
	if err != nil {
		h.c.Log.WithField("error", err).Error("Verification error")
		return nil
	}
	if action == verification.ClientDrop {
		return nil
	}
	registered := (action == verification.ClientRegistered)

	network := c.searchNetworksFor(reqIP)
	if network == nil {
		h.c.Log.WithFields(verbose.Fields{
			"ip":         reqIP.String(),
			"registered": registered,
		}).Notice("Got a RELEASE for IP not in a scope")
		return nil
	}

	lease, _ := network.getLeaseByIP(reqIP, registered)
	h.c.Log.Debugf("%#v", lease)
	if lease == nil || !bytes.Equal(lease.MAC, p.CHAddr()) {
		leaseMac := ""
		if lease != nil {
			leaseMac = lease.MAC.String()
		}

		h.c.Log.WithFields(verbose.Fields{
			"ip":         reqIP.String(),
			"mac":        p.CHAddr().String(),
			"lease_mac":  leaseMac,
			"network":    network.name,
			"registered": registered,
		}).Notice("Client tried to release lease not belonging to them")
		return nil
	}
	h.c.Log.WithFields(verbose.Fields{
		"ip":         lease.IP.String(),
		"mac":        lease.MAC.String(),
		"network":    network.name,
		"registered": registered,
		"action":     "release",
	}).Info("Releasing lease")
	lease.Start = time.Unix(1, 0)
	lease.End = time.Unix(1, 0)
	if err := h.c.Store.PutLease(lease); err != nil {
		h.c.Log.WithFields(verbose.Fields{
			"mac":   p.CHAddr().String(),
			"error": err,
		}).Error("Error saving lease")
	}
	return nil
}

// Handle DHCP DECLINE messages
func (h *Handler) handleDecline(p dhcp4.Packet, options dhcp4.Options) dhcp4.Packet {
	reqIP := p.CIAddr()
	if reqIP == nil || reqIP.Equal(net.IPv4zero) {
		return nil
	}

	action, err := h.c.Verification.VerifyClient(p.CHAddr())
	if err != nil {
		h.c.Log.WithField("error", err).Error("Verification error")
		return nil
	}
	if action == verification.ClientDrop {
		return nil
	}
	registered := (action == verification.ClientRegistered)

	network := c.searchNetworksFor(reqIP)
	if network == nil {
		h.c.Log.WithFields(verbose.Fields{
			"ip":         reqIP.String(),
			"registered": registered,
		}).Notice("Got a DECLINE for IP not in a scope")
		return nil
	}

	lease, _ := network.getLeaseByIP(reqIP, registered)
	if lease == nil || !bytes.Equal(lease.MAC, p.CHAddr()) {
		leaseMac := ""
		if lease != nil {
			leaseMac = lease.MAC.String()
		}

		h.c.Log.WithFields(verbose.Fields{
			"declined_ip": reqIP.String(),
			"mac":         p.CHAddr().String(),
			"lease_mac":   leaseMac,
			"network":     network.name,
			"registered":  registered,
		}).Notice("Client tried to decline lease not belonging to them")
		return nil
	}
	h.c.Log.WithFields(verbose.Fields{
		"ip":         lease.IP.String(),
		"mac":        lease.MAC.String(),
		"network":    network.name,
		"registered": registered,
		"action":     "decline",
	}).Notice("Abandoned lease")
	lease.IsAbandoned = true
	lease.Start = time.Unix(1, 0)
	lease.End = time.Unix(1, 0)
	if err := h.c.Store.PutLease(lease); err != nil {
		h.c.Log.WithFields(verbose.Fields{
			"mac":   p.CHAddr().String(),
			"error": err,
		}).Error("Error saving lease")
	}
	return nil
}

func (h *Handler) handleInform(p dhcp4.Packet, options dhcp4.Options) dhcp4.Packet {
	ip := p.CIAddr()
	if ip == nil || ip.Equal(net.IPv4zero) {
		return nil
	}

	network := c.searchNetworksFor(ip)
	if network == nil {
		return nil
	}

	pool := network.getPoolOfIP(ip)
	if pool == nil {
		return nil
	}

	action, err := h.c.Verification.VerifyClient(p.CHAddr())
	if err != nil {
		h.c.Log.WithField("error", err).Error("Verification error")
		return nil
	}
	if action == verification.ClientDrop {
		return nil
	}
	registered := (action == verification.ClientRegistered)

	leaseOptions := pool.getOptions(registered)

	h.c.Log.WithFields(verbose.Fields{
		"ip":     ip.String(),
		"mac":    p.CHAddr().String(),
		"action": "inform",
	}).Info("Informing client")

	return dhcp4.ReplyPacket(
		p,
		dhcp4.ACK,
		c.global.serverIdentifier,
		net.IP([]byte{0, 0, 0, 0}),
		0,
		leaseOptions.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]),
	)
}
