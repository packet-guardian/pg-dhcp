package server

import (
	"os"

	"github.com/packet-guardian/pg-dhcp/store"
)

func setUpLeaseStore() (*store.Store, error) {
	return store.NewStore("testing.db")
}

func tearDownLeaseStore(db *store.Store) {
	db.Close()
	os.Remove("testing.db")
}
