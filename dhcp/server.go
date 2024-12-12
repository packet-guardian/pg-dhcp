// Original implementation: 2014 Skagerrak Software - http://www.skagerraksoftware.com/
// Modifications: 2017 Lee Keitel

package dhcp4

import (
	"fmt"
	"net"
	"strconv"

	"golang.org/x/net/ipv4"
)

// A Handler takes a DHCP request packet and generates a response to the client
type Handler interface {
	ServeDHCP(req Packet, msgType MessageType, options Options) Packet
}

// ServeConn is the bare minimum connection functions required by Serve()
// It allows you to create custom connections for greater control,
// such as ServeIfConn (see serverif.go), which locks to a given interface.
// type ServeConn interface {
// 	ReadFrom(b []byte) (n int, addr net.Addr, err error)
// 	WriteTo(b []byte, addr net.Addr) (n int, err error)
// }

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
// can provide a workaround to this problem.
func Serve(conn net.PacketConn, handler Handler, workers int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
			return
		}
	}()

	taskQueue := startWorkers(workers, conn, handler)

	cmconn := ipv4.NewPacketConn(conn)
	if err := cmconn.SetControlMessage(ipv4.FlagInterface, true); err != nil {
		return err
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	ifaceAddrs := make([]net.IP, len(ifaces))
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			host, _, _ := net.ParseCIDR(addr.String())
			if host.To4() != nil {
				ifaceAddrs[iface.Index-1] = host
				break
			}
		}
	}

	for {
		buffer := make([]byte, 1500)
		n, cm, addr, err := cmconn.ReadFrom(buffer)
		if err != nil {
			close(taskQueue)
			return err
		}
		if n < 240 { // Packet too small to be DHCP
			continue
		}
		req := Packet(buffer[:n])
		if req.HLen() > 16 { // Invalid size
			continue
		}

		select {
		case taskQueue <- job{packet: req, dst: ifaceAddrs[cm.IfIndex-1], from: addr}:
		default:
			fmt.Println("Task queue full")
		}
	}
}

func process(conn net.PacketConn, handler Handler, j *job) {
	options := j.packet.ParseOptions()

	t := options[OptionDHCPMessageType]
	if len(t) != 1 {
		return
	}

	reqType := MessageType(t[0])
	if reqType < Discover || reqType > Inform {
		return
	}

	// If there's no DHCP relay, use the local server address as if it was the
	// gateway for processing
	if j.packet.GIAddr().IsUnspecified() {
		j.packet.SetGIAddr(j.dst)
	}

	if res := handler.ServeDHCP(j.packet, reqType, options); res != nil {
		// If coming from a relay and the relay address is not
		// a local server address, unicast back
		if !j.packet.GIAddr().Equal(j.dst) && !j.packet.GIAddr().IsUnspecified() {
			if _, e := conn.WriteTo(res, j.from); e != nil {
				panic(e)
			}
			return
		}

		ipStr, portStr, err := net.SplitHostPort(j.from.String())
		if err != nil {
			return
		}

		// If IP not available or broadcast bit is set, broadcast
		if net.ParseIP(ipStr).IsUnspecified() || j.packet.Broadcast() {
			port, _ := strconv.Atoi(portStr)
			j.from = &net.UDPAddr{IP: net.IPv4bcast, Port: port}
		}
		if _, e := conn.WriteTo(res, j.from); e != nil {
			panic(e)
		}
	}
}

type job struct {
	packet Packet
	dst    net.IP
	from   net.Addr
}

func startWorkers(num int, conn net.PacketConn, handler Handler) chan job {
	tasks := make(chan job, num*2)

	for i := 1; i <= num; i++ {
		fmt.Printf("Starting worker %d\n", i)
		go worker(conn, handler, tasks)
	}

	return tasks
}

func worker(conn net.PacketConn, handler Handler, tasks <-chan job) {
	for j := range tasks {
		process(conn, handler, &j)
	}
	fmt.Println("Worker stopping")
}
