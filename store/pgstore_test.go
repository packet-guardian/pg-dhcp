//+build mysql

package store

import (
	"testing"
)

func setUpPGStore() (*PGStore, error) {
	s, err := NewPGStore(mysqlCfg, "lease", "device", "blacklist")
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec("DROP TABLE IF EXISTS lease, device, blacklist")
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(`CREATE TABLE "device" (
		"id" INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
		"mac" VARCHAR(17) NOT NULL UNIQUE KEY
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1`)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(`CREATE TABLE "blacklist" (
		"id" INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
		"value" VARCHAR(255) NOT NULL UNIQUE KEY
	) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1`)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return s, err
}

func tearDownPGStore(s *PGStore) {
	s.db.Exec("DROP TABLE IF EXISTS lease, device, blacklist")
	s.Close()
}

func TestLeasePGStore(t *testing.T) {
	if !mysqlAvailable {
		t.Skipf("MySQL server not running on %s", mysqlCfg.Addr)
	}

	store, err := setUpPGStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownPGStore(store)
	testLeaseStore(t, store)
}

func TestForEachLeasePGStore(t *testing.T) {
	if !mysqlAvailable {
		t.Skipf("MySQL server not running on %s", mysqlCfg.Addr)
	}

	store, err := setUpPGStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownPGStore(store)
	testForEachLease(t, store)
}

func TestDevicePGStore(t *testing.T) {
	if !mysqlAvailable {
		t.Skipf("MySQL server not running on %s", mysqlCfg.Addr)
	}

	store, err := setUpPGStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownPGStore(store)

	// The device table is managed by a separate application so we need to seed it with data
	_, err = store.db.Exec(`INSERT INTO "device" ("mac") VALUES ('12:34:56:ab:cd:ee'), ('12:34:56:ab:cd:ef')`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.db.Exec(`INSERT INTO "blacklist" ("value") VALUES ('12:34:56:ab:cd:ef')`)
	if err != nil {
		t.Fatal(err)
	}

	testDeviceStore(t, store)
}

func TestDeviceStoreNonExistantDevicePGStore(t *testing.T) {
	if !mysqlAvailable {
		t.Skipf("MySQL server not running on %s", mysqlCfg.Addr)
	}

	store, err := setUpPGStore()
	if err != nil {
		t.Fatal(err)
	}
	defer tearDownPGStore(store)
	testDeviceStoreNonExistantDevice(t, store)
}
