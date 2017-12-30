package server

import (
	"github.com/packet-guardian/pg-dhcp/store"
)

func setUpStore() (store.Store, error) {
	return store.NewMemoryStore()
}

func tearDownStore(db store.Store) {
	db.Close()
}
