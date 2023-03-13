package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/packet-guardian/pg-dhcp/models"
	"github.com/packet-guardian/pg-dhcp/store"
)

var (
	source_dsn = ""
	target_dsn = ""
)

func init() {
	flag.StringVar(&source_dsn, "src", "", "Data Source")
	flag.StringVar(&target_dsn, "dst", "", "Data Target")
}

func main() {
	flag.Parse()

	src_store, err := getStore(source_dsn)
	if err != nil {
		log.Fatalf("Error opening store: %s", err.Error())
	}
	defer src_store.Close()

	dst_store, err := getStore(target_dsn)
	if err != nil {
		log.Fatalf("Error opening store: %s", err.Error())
	}
	defer dst_store.Close()

	if err := migrateStores(src_store, dst_store); err != nil {
		log.Fatalf("Error opening store: %s", err.Error())
	}
}

func getStore(dsn string) (store.Store, error) {
	// DSN: type;username:password@address
	// MySQL: mysql;user:pass@127.0.0.1:3306/db
	// PG: pg;user:pass@127.0.0.1:3306/db
	// Bolt: bolt@filepath.db

	split_one := strings.Split(dsn, "@")
	if len(split_one) != 2 {
		return nil, fmt.Errorf("error in DSN: %s", dsn)
	}

	address := split_one[1]

	if split_one[0] == "bolt" {
		return store.NewBoltStore(address)
	}

	dbname_split := strings.Split(address, "/")
	if len(dbname_split) != 2 {
		return nil, fmt.Errorf("error in DSN: %s", dsn)
	}

	address = dbname_split[0]
	dbname := dbname_split[1]

	split_two := strings.Split(split_one[0], ";")
	if len(split_two) != 2 {
		return nil, fmt.Errorf("error in DSN: %s", dsn)
	}

	data_type := split_two[0]

	split_three := strings.Split(split_two[1], ":")
	if len(split_three) != 2 {
		return nil, fmt.Errorf("error in DSN: %s", dsn)
	}

	username := split_three[0]
	password := split_three[1]
	netAddr := fmt.Sprintf("tcp(%s)", address)

	mysql_config := mysql.NewConfig()
	mysql_config.User = username
	mysql_config.Passwd = password
	mysql_config.Addr = netAddr
	mysql_config.DBName = dbname
	mysql_config.Timeout = 30 * time.Second

	if data_type == "mysql" {
		return store.NewMySQLStore(mysql_config, "lease", "device")
	} else if data_type == "pg" {
		return store.NewPGStore(mysql_config, "lease", "device", "blacklist")
	}

	return nil, fmt.Errorf("unsupported data store: %s", data_type)
}

func migrateStores(src, dst store.Store) error {
	if err := migrateLeases(src, dst); err != nil {
		return err
	}

	return migrateDevices(src, dst)
}

func migrateLeases(src, dst store.Store) error {
	count := 0

	src.ForEachLease(func(_ *models.Lease) {
		count++
	})

	log.Printf("%d leases found, migrating leases...", count)

	var err error

	src.ForEachLease(func(lease *models.Lease) {
		if err = dst.PutLease(lease); err != nil {
			return
		}
	})

	return err
}

func migrateDevices(src, dst store.Store) error {
	count := 0

	src.ForEachDevice(func(_ *models.Device) {
		count++
	})

	log.Printf("%d devices found, migrating devices...", count)

	var err error

	src.ForEachDevice(func(device *models.Device) {
		if err = dst.PutDevice(device); err != nil {
			return
		}
	})

	return err
}
