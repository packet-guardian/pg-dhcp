// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO: Clean up the handler functions. There's a lot of duplicated code that
// could be extracted to a function.

package server

import (
	"bytes"
	"errors"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/lfkeitel/verbose/v5"
	dhcp4 "github.com/packet-guardian/pg-dhcp/dhcp"
	"github.com/packet-guardian/pg-dhcp/models"
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
	closing      bool
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
	logger := verbose.New()

	// Add standard output handler
	sh := verbose.NewTextTransport()
	logger.AddTransport(sh)
	sh.SetMinLevel(verbose.LogLevelInfo)

	return logger
}

// ListenAndServe starts the DHCP Handler listening on port 67 for packets.
// This is blocking like HTTP's ListenAndServe method.
func (h *Handler) ListenAndServe() error {
	if h.c.Workers <= 0 {
		return errors.New("Server.Workers needs to be greater than 0")
	}

	h.c.Log.Info("Starting DHCP server...")
	l, err := net.ListenPacket("udp4", ":67")
	if err != nil {
		return err
	}
	h.conn = l
	err = dhcp4.Serve(l, h, h.c.Workers)
	if h.closing {
		return nil
	}
	return err
}

func (h *Handler) Close() error {
	h.closing = true
	h.conn.Close()
	h.c.Store.Close()
	return nil
}

// LoadLeases will import any current leases saved to the database.
func (h *Handler) LoadLeases() error {
	h.c.Store.ForEachLease(func(l *models.Lease) {
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
		}).Debug("Incoming request")
	}

	device, err := h.c.Store.GetDevice(p.CHAddr())
	if err != nil {
		h.c.Log.WithField("error", err.Error()).Error("Failed getting device")
		return nil
	}
	if device.Blacklisted && h.c.BlockBlacklist {
		return nil
	}

	var response dhcp4.Packet
	switch msgType {
	case dhcp4.Discover:
		response = h.handleDiscover(p, options, device)
	case dhcp4.Request:
		response = h.handleRequest(p, options, device)
	case dhcp4.Release:
		response = h.handleRelease(p, options, device)
	case dhcp4.Decline:
		response = h.handleDecline(p, options, device)
	case dhcp4.Inform:
		response = h.handleInform(p, options, device)
	}
	return response
}

func isDeviceRegistered(d *models.Device) bool {
	return d.Registered && !d.Blacklisted
}

// Handle DHCP DISCOVER messages
func (h *Handler) handleDiscover(p dhcp4.Packet, options dhcp4.Options, device *models.Device) dhcp4.Packet {
	start := time.Now()

	gatewayIP := p.GIAddr().String()
	// Get network object that the relay IP belongs to
	h.gatewayMutex.Lock()
	network, ok := h.gatewayCache[gatewayIP]
	if !ok {
		// That gateway hasn't been seen before, find its network
		network = c.searchNetworksFor(p.GIAddr())
		if network == nil {
			h.gatewayMutex.Unlock()
			h.c.Log.WithField("relay_ip", gatewayIP).Notice("Network not found")
			return nil
		}
		// Add to cache for later
		h.gatewayCache[gatewayIP] = network
	}
	network.Lock()
	defer network.Unlock()
	h.gatewayMutex.Unlock()

	registered := isDeviceRegistered(device) && !network.ignoreRegistration

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
				"mac":        p.CHAddr().String(),
			}).Alert("No free leases available in network")
			return nil
		}
	}

	// Set temporary offered flag and end time
	lease.Offered = true
	lease.Start = time.Now()
	lease.End = time.Now().Add(time.Duration(30) * time.Second) // Set a short end time so it's not offered to other clients
	lease.MAC = make([]byte, len(p.CHAddr()))
	copy(lease.MAC, p.CHAddr())
	// No Save because this is a temporary "lease", if the client accepts then we commit to storage
	// Get options
	leaseOptions := pool.getOptions(registered)

	h.c.Log.WithFields(verbose.Fields{
		"ip":         lease.IP.String(),
		"mac":        p.CHAddr().String(),
		"registered": registered,
		"network":    network.name,
		"action":     "offer",
		"took":       time.Since(start).String(),
		"relay_ip":   gatewayIP,
	}).Info("Offering lease to client")

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
func (h *Handler) handleRequest(p dhcp4.Packet, options dhcp4.Options, device *models.Device) dhcp4.Packet {
	if server, ok := options[dhcp4.OptionServerIdentifier]; ok && !net.IP(server).Equal(c.global.serverIdentifier) {
		return nil // Message not for this dhcp server
	}

	start := time.Now()
	reqIP := net.IP(options[dhcp4.OptionRequestedIPAddress])
	if reqIP == nil {
		reqIP = net.IP(p.CIAddr())
	}

	if len(reqIP) != 4 || reqIP.Equal(net.IPv4zero) {
		return dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil)
	}

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

	registered := isDeviceRegistered(device)

	if network == nil {
		h.c.Log.WithFields(verbose.Fields{
			"ip":         reqIP.String(),
			"registered": registered,
		}).Info("Got a REQUEST for IP not in a scope")
		return dhcp4.ReplyPacket(p, dhcp4.NAK, c.global.serverIdentifier, nil, 0, nil)
	}
	network.Lock()
	defer network.Unlock()

	registered = registered && !network.ignoreRegistration

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
	} else {
		lease.Hostname = ""
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
		"ip":          lease.IP.String(),
		"mac":         lease.MAC.String(),
		"duration":    leaseDur.String(),
		"network":     network.name,
		"relay_ip":    p.GIAddr().String(),
		"registered":  device.Registered,
		"hostname":    lease.Hostname,
		"action":      "request_ack",
		"blacklisted": device.Blacklisted,
		"took":        time.Since(start).String(),
	}).Info("Acknowledging request")

	if device.Registered {
		device.LastSeen = time.Now()
		if err := h.c.Store.PutDevice(device); err != nil {
			// We won't consider this a critical error, still give out the lease
			h.c.Log.WithField("Err", err).Error("Failed updating device last_seen attribute")
		}
	}

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
func (h *Handler) handleRelease(p dhcp4.Packet, options dhcp4.Options, device *models.Device) dhcp4.Packet {
	start := time.Now()
	reqIP := p.CIAddr()
	if reqIP == nil || reqIP.Equal(net.IPv4zero) {
		return nil
	}

	registered := isDeviceRegistered(device)

	network := c.searchNetworksFor(reqIP)
	if network == nil {
		h.c.Log.WithFields(verbose.Fields{
			"ip":         reqIP.String(),
			"registered": registered,
		}).Notice("Got a RELEASE for IP not in a scope")
		return nil
	}
	network.Lock()
	defer network.Unlock()

	registered = registered && !network.ignoreRegistration

	lease, _ := network.getLeaseByIP(reqIP, registered)
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
		"relay_ip":   p.GIAddr().String(),
		"registered": device.Registered,
		"action":     "release",
		"took":       time.Since(start).String(),
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
// TODO: Decline would never work because the ciaddr field will always be 0
// for a properly formed DECLINE message. Also, a DECLINE has nothing to do
// with a client being registered or not.
func (h *Handler) handleDecline(p dhcp4.Packet, options dhcp4.Options, device *models.Device) dhcp4.Packet {
	start := time.Now()
	reqIP := p.CIAddr()
	if reqIP == nil || reqIP.Equal(net.IPv4zero) {
		return nil
	}

	registered := isDeviceRegistered(device)

	network := c.searchNetworksFor(reqIP)
	if network == nil {
		h.c.Log.WithFields(verbose.Fields{
			"ip":         reqIP.String(),
			"registered": registered,
		}).Notice("Got a DECLINE for IP not in a scope")
		return nil
	}
	network.Lock()
	defer network.Unlock()

	registered = registered && !network.ignoreRegistration

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
		"relay_ip":   p.GIAddr().String(),
		"registered": device.Registered,
		"action":     "decline",
		"took":       time.Since(start).String(),
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

func (h *Handler) handleInform(p dhcp4.Packet, options dhcp4.Options, device *models.Device) dhcp4.Packet {
	start := time.Now()
	ip := p.CIAddr()
	if ip == nil || ip.Equal(net.IPv4zero) {
		return nil
	}

	network := c.searchNetworksFor(ip)
	if network == nil {
		return nil
	}
	network.Lock()
	defer network.Unlock()

	pool := network.getPoolOfIP(ip)
	if pool == nil {
		return nil
	}

	registered := isDeviceRegistered(device) && !network.ignoreRegistration

	leaseOptions := pool.getOptions(registered)

	h.c.Log.WithFields(verbose.Fields{
		"ip":       ip.String(),
		"mac":      p.CHAddr().String(),
		"network":  network.name,
		"relay_ip": p.GIAddr().String(),
		"action":   "inform",
		"took":     time.Since(start).String(),
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
