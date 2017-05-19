package dhcp4

import (
	"net"
	"strconv"
)

type Handler interface {
	ServeDHCP(req Packet, msgType MessageType, options Options) Packet
}

// ServeConn is the bare minimum connection functions required by Serve()
// It allows you to create custom connections for greater control,
// such as ServeIfConn (see serverif.go), which locks to a given interface.
type ServeConn interface {
	ReadFrom(b []byte) (n int, addr net.Addr, err error)
	WriteTo(b []byte, addr net.Addr) (n int, err error)
}

// ListenAndServe listens on the UDP network address addr and then calls
// Serve with handler to handle requests on incoming packets.
func ListenAndServe(handler Handler) error {
	l, err := net.ListenPacket("udp4", ":67")
	if err != nil {
		return err
	}
	defer l.Close()
	return Serve(l, handler)
}

// Serve takes a ServeConn (such as a net.PacketConn) that it uses for both
// reading and writing DHCP packets. Every packet is passed to the handler,
// which processes it and optionally return a response packet for writing back
// to the network.
//
// To capture limited broadcast packets (sent to 255.255.255.255), you must
// listen on a socket bound to IP_ADDRANY (0.0.0.0). This means that broadcast
// packets sent to any interface on the system may be delivered to this
// socket.  See: https://code.google.com/p/go/issues/detail?id=7106
//
// Additionally, response packets may not return to the same
// interface that the request was received from.  Writing a custom ServeConn,
// or using ServeIf() can provide a workaround to this problem.
func Serve(conn ServeConn, handler Handler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
			return
		}
	}()

	for {
		buffer := make([]byte, 1500)
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			return err
		}
		if n < 240 { // Packet too small to be DHCP
			continue
		}
		req := Packet(buffer[:n])
		if req.HLen() > 16 { // Invalid size
			continue
		}
		process(&request{
			conn:    conn,
			p:       req,
			handler: handler,
			from:    addr,
		})
	}
}

type request struct {
	conn    ServeConn
	p       Packet
	handler Handler
	from    net.Addr
}

func process(req *request) {
	options := req.p.ParseOptions()

	t := options[OptionDHCPMessageType]
	if len(t) != 1 {
		return
	}

	reqType := MessageType(t[0])
	if reqType < Discover || reqType > Inform {
		return
	}

	if res := req.handler.ServeDHCP(req.p, reqType, options); res != nil {
		// If coming from a relay, unicast back
		if !req.p.GIAddr().Equal(net.IPv4zero) {
			if _, e := req.conn.WriteTo(res, req.from); e != nil {
				panic(e)
			}
			return
		}

		ipStr, portStr, err := net.SplitHostPort(req.from.String())
		if err != nil {
			return
		}

		// If IP not available or broadcast bit is set, broadcast
		if net.ParseIP(ipStr).Equal(net.IPv4zero) || req.p.Broadcast() {
			port, _ := strconv.Atoi(portStr)
			req.from = &net.UDPAddr{IP: net.IPv4bcast, Port: port}
		}
		if _, e := req.conn.WriteTo(res, req.from); e != nil {
			panic(e)
		}
	}
}
