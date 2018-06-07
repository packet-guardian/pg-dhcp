package store

import (
	"os"
	"testing"
)

func setUpBoltDBStore() (*BoltStore, error) {
	return NewBoltStore("test.db")
}

func tearDownBoltDBStore(db *BoltStore) {
	db.Close()
	os.Remove("test.db")
}

func TestLeaseBoltDBStore(t *testing.T) {
	store, err := setUpBoltDBStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownBoltDBStore(store)
	testLeaseStore(t, store)
}

func TestForEachLeaseBoltDBStore(t *testing.T) {
	store, err := setUpBoltDBStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownBoltDBStore(store)
	testForEachLease(t, store)
}

func TestDeviceBoltDBStore(t *testing.T) {
	store, err := setUpBoltDBStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownBoltDBStore(store)
	testDeviceStore(t, store)
}

func TestDeviceStoreNonExistantDeviceBoltDBStore(t *testing.T) {
	store, err := setUpBoltDBStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownBoltDBStore(store)
	testDeviceStoreNonExistantDevice(t, store)
}

func TestForEachDeviceBoltDBStore(t *testing.T) {
	store, err := setUpBoltDBStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownBoltDBStore(store)
	testForEachDevice(t, store)
}
