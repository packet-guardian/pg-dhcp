package store

import (
	"os"
	"reflect"
	"testing"
)

func setUpLeaseStore() (*Store, error) {
	return NewStore("leases_test.db")
}

func tearDownLeaseStore(db *Store) {
	db.Close()
	os.Remove("leases_test.db")
}

func TestLeaseStore(t *testing.T) {
	store, err := setUpLeaseStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownLeaseStore(store)

	lease := leaseTests[0].actual
	if err := store.PutLease(lease); err != nil {
		t.Fatal(err)
	}

	lease2, err := store.GetLease(lease.IP)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(lease, lease2) {
		t.Fatalf("Leases don't match")
	}
}

func TestForEachLease(t *testing.T) {
	store, err := setUpLeaseStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownLeaseStore(store)

	lease1 := leaseTests[0].actual
	lease2 := leaseTests[1].actual

	store.PutLease(lease1)
	store.PutLease(lease2)

	var newLease1, newLease2 *Lease

	store.ForEachLease(func(l *Lease) {
		if l.IP.String() == "10.0.2.5" {
			newLease1 = l
		} else if l.IP.String() == "10.0.2.6" {
			newLease2 = l
		}
	})

	if newLease1 == nil {
		t.Error("newLease1 is nil")
	}
	if newLease2 == nil {
		t.Error("newLease2 is nil")
	}
}
