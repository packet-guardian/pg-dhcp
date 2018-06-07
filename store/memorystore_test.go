package store

import (
	"testing"
)

func setUpMemoryStore() (*MemoryStore, error) {
	return NewMemoryStore()
}

func TestLeaseMemoryStore(t *testing.T) {
	store, err := setUpMemoryStore()
	if err != nil {
		t.Fatal(err)
	}
	testLeaseStore(t, store)
}

func TestForEachLeaseMemoryStore(t *testing.T) {
	store, err := setUpMemoryStore()
	if err != nil {
		t.Fatal(err)
	}
	testForEachLease(t, store)
}

func TestDeviceMemoryStore(t *testing.T) {
	store, err := setUpMemoryStore()
	if err != nil {
		t.Fatal(err)
	}
	testDeviceStore(t, store)
}

func TestDeviceStoreNonExistantDeviceMemoryStore(t *testing.T) {
	store, err := setUpMemoryStore()
	if err != nil {
		t.Fatal(err)
	}
	testDeviceStoreNonExistantDevice(t, store)
}

func TestForEachDeviceMemoryStore(t *testing.T) {
	store, err := setUpMemoryStore()
	if err != nil {
		t.Fatal(err)
	}
	testForEachDevice(t, store)
}
