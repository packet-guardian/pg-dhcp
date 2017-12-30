package server

import (
	"os"

	"github.com/packet-guardian/pg-dhcp/store"
)

func setUpStore() (store.Store, error) {
	os.Remove("testing.db")
	return store.NewBoltStore("testing.db")
}

func tearDownStore(db store.Store) {
	db.Close()
	os.Remove("testing.db")
}
