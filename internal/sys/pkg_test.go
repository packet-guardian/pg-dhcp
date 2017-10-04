package sys

import (
	"os"

	"github.com/packet-guardian/pg-dhcp/store"
)

func setUpLeaseStore() (*store.Store, error) {
	os.Remove("testing.db")
	return store.NewStore("testing.db")
}

func tearDownLeaseStore(db *store.Store) {
	db.Close()
	os.Remove("testing.db")
}
