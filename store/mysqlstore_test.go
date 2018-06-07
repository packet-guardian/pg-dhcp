//+build mysql

package store

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/go-sql-driver/mysql"
)

var (
	mysqlCfg       *mysql.Config
	mysqlAvailable bool
)

func init() {
	// get environment variables
	env := func(key, defaultValue string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return defaultValue
	}

	user := env("MYSQL_TEST_USER", "root")
	pass := env("MYSQL_TEST_PASS", "password")
	prot := env("MYSQL_TEST_PROT", "tcp")
	addr := env("MYSQL_TEST_ADDR", "localhost:3306")
	dbname := env("MYSQL_TEST_DBNAME", "gotest")
	netAddr := fmt.Sprintf("%s(%s)", prot, addr)
	dsn := fmt.Sprintf("%s:%s@%s/%s?timeout=30s", user, pass, netAddr, dbname)
	c, err := net.Dial(prot, addr)
	if err == nil {
		mysqlAvailable = true
		mysqlCfg, _ = mysql.ParseDSN(dsn)
		c.Close()
	}
}

func setUpMySQLStore(t *testing.T) (*MySQLStore, error) {
	s, err := NewMySQLStore(mysqlCfg, "lease", "device")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.db.Exec("DROP TABLE IF EXISTS lease, device")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.db.Exec(`CREATE TABLE "device" (
		"mac" VARCHAR(17) NOT NULL UNIQUE KEY,
		"registered" TINYINT DEFAULT 0,
		"blacklisted" TINYINT DEFAULT 0
	) ENGINE=InnoDB DEFAULT CHARSET=utf8`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.db.Exec(`CREATE TABLE "lease" (
		"ip" VARCHAR(15) NOT NULL UNIQUE KEY,
		"mac" VARCHAR(17) NOT NULL,
		"network" TEXT NOT NULL,
		"start" INTEGER NOT NULL,
		"end" INTEGER NOT NULL,
		"hostname" TEXT NOT NULL,
		"abandoned" TINYINT DEFAULT 0,
		"registered" TINYINT DEFAULT 0
	) ENGINE=InnoDB DEFAULT CHARSET=utf8`)
	if err != nil {
		t.Fatal(err)
	}
	return s, err
}

func tearDownMySQLStore(s *MySQLStore) {
	s.db.Exec("DROP TABLE IF EXISTS lease, device")
	s.Close()
}

func TestLeaseMySQLStore(t *testing.T) {
	if !mysqlAvailable {
		t.Skipf("MySQL server not running on %s", mysqlCfg.Addr)
	}

	store, err := setUpMySQLStore(t)
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownMySQLStore(store)
	testLeaseStore(t, store)
}

func TestForEachLeaseMySQLStore(t *testing.T) {
	if !mysqlAvailable {
		t.Skipf("MySQL server not running on %s", mysqlCfg.Addr)
	}

	store, err := setUpMySQLStore(t)
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownMySQLStore(store)
	testForEachLease(t, store)
}

func TestDeviceMySQLStore(t *testing.T) {
	if !mysqlAvailable {
		t.Skipf("MySQL server not running on %s", mysqlCfg.Addr)
	}

	store, err := setUpMySQLStore(t)
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownMySQLStore(store)
	testDeviceStore(t, store)
}

func TestDeviceStoreNonExistantDeviceMySQLStore(t *testing.T) {
	if !mysqlAvailable {
		t.Skipf("MySQL server not running on %s", mysqlCfg.Addr)
	}

	store, err := setUpMySQLStore(t)
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownMySQLStore(store)
	testDeviceStoreNonExistantDevice(t, store)
}

func TestForEachDeviceMySQLStore(t *testing.T) {
	if !mysqlAvailable {
		t.Skipf("MySQL server not running on %s", mysqlCfg.Addr)
	}

	store, err := setUpMySQLStore(t)
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownMySQLStore(store)
	testForEachDevice(t, store)
}
