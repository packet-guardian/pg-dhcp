package management

import (
	"os"
	"sort"

	"github.com/lfkeitel/verbose"
	"github.com/packet-guardian/pg-dhcp/internal/server"
	"github.com/packet-guardian/pg-dhcp/store"
)

type fatalLogger interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

func setUpTest(t fatalLogger) (*server.Handler, *store.Store) {
	db, err := setUpStore()
	if err != nil {
		t.Fatal(err)
	}

	// Setup Configuration
	c, err := server.ParseFile("../internal/server/testdata/testConfig.conf")
	if err != nil {
		t.Fatalf("Test config failed parsing: %v", err)
	}

	sc := &server.ServerConfig{
		Env:   server.EnvTesting,
		Log:   verbose.New(""),
		Store: db,
	}

	return server.NewDHCPServer(c, sc), db
}

func setUpStore() (*store.Store, error) {
	os.Remove("testing.db")
	return store.NewStore("testing.db")
}

func tearDownStore(db *store.Store) {
	db.Close()
	os.Remove("testing.db")
}

func stringSliceEqual(a, b []string) bool {
	sortString(a)
	sortString(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sortString(a []string) {
	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
}
