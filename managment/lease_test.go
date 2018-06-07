package management

import (
	"net"
	"testing"

	"github.com/packet-guardian/pg-dhcp/models"
)

func TestLeaseGetLeaseRPC(t *testing.T) {
	_, db := setUpTest(t)
	defer tearDownStore(db)

	mac1 := net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})
	ip1 := net.ParseIP("10.0.2.1")

	mac2 := net.HardwareAddr([]byte{0x22, 0x34, 0x56, 0xab, 0xcd, 0xef})
	ip2 := net.ParseIP("10.0.2.2")

	ip3 := net.ParseIP("10.0.2.3")

	db.PutLease(&models.Lease{
		MAC: mac1,
		IP:  ip1,
	})
	db.PutLease(&models.Lease{
		MAC: mac2,
		IP:  ip2,
	})

	rpc := &Lease{store: db}

	lease := new(models.Lease)
	if err := rpc.Get(ip2, lease); err != nil {
		t.Fatal(err)
	}
	if (lease).MAC.String() != mac2.String() {
		t.Fatalf("Wrong lease returned. Expected %s, got %s", mac2.String(), (lease).MAC.String())
	}

	if err := rpc.Get(ip1, lease); err != nil {
		t.Fatal(err)
	}
	if (lease).MAC.String() != mac1.String() {
		t.Fatalf("Wrong lease returned. Expected %s, got %s", mac1.String(), (lease).MAC.String())
	}

	lease = new(models.Lease)
	if err := rpc.Get(ip3, lease); err != nil {
		t.Fatal(err)
	}
	if lease.IP != nil {
		t.Fatalf("Non existant lease returned. Got %s", lease.IP.String())
	}
}

func TestLeaseGetAllFromNetworkRPC(t *testing.T) {
	handler, db := setUpTest(t)
	defer tearDownStore(db)

	mac1 := net.HardwareAddr([]byte{0x12, 0x34, 0x56, 0xab, 0xcd, 0xef})
	ip1 := net.ParseIP("10.0.2.10")

	mac2 := net.HardwareAddr([]byte{0x22, 0x34, 0x56, 0xab, 0xcd, 0xef})
	ip2 := net.ParseIP("10.0.2.11")

	mac3 := net.HardwareAddr([]byte{0x32, 0x34, 0x56, 0xab, 0xcd, 0xef})
	ip3 := net.ParseIP("10.0.2.12")

	// Generate and populate store with leases on network1
	db.PutLease(&models.Lease{
		MAC:     mac1,
		IP:      ip1,
		Network: "network1",
	})
	db.PutLease(&models.Lease{
		MAC:     mac2,
		IP:      ip2,
		Network: "network1",
	})
	db.PutLease(&models.Lease{
		MAC:     mac3,
		IP:      ip3,
		Network: "network1",
	})

	// Make server load generated leases into network
	if err := handler.LoadLeases(); err != nil {
		t.Fatal(err)
	}

	rpc := &Lease{store: db}

	var leases []*models.Lease
	if err := rpc.GetAllFromNetwork("network1", &leases); err != nil {
		t.Fatal(err)
	}
	if len(leases) != 3 {
		t.Fatalf("Incorrect number of leases. Expected 3, got %d", len(leases))
	}
}
