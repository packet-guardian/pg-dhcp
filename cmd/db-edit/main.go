package main

import (
	"encoding/csv"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/packet-guardian/pg-dhcp/store"
)

var (
	dbPath    string
	inputFile string

	leaseBucket = []byte("leases")
)

func init() {
	flag.StringVar(&dbPath, "db", "database.db", "BoltDB database file")
	flag.StringVar(&inputFile, "in", "input.csv", "Input data in CSV format")
}

func main() {
	flag.Parse()

	leases, err := parseLeases()
	if err != nil {
		log.Fatal(err)
	}

	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Batch(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(leaseBucket)
		if err != nil {
			return err
		}

		lb := tx.Bucket(leaseBucket)

		for _, lease := range leases {
			log.Printf("Loading %s\n", lease.IP)
			if err := lb.Put([]byte(lease.IP.To4()), lease.Serialize()); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func parseLeases() ([]*store.Lease, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = 10
	leases := make([]*store.Lease, 0, 10)

	// Discard header
	_, err = csvReader.Read()
	if err != nil {
		return nil, err
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// IP,MAC,Network,Start,End,Hostname,IsAbandoned,Offered,Registered,Used
		l := &store.Lease{}
		l.IP = net.ParseIP(record[0])
		l.Network = record[2]
		l.Hostname = record[5]
		l.IsAbandoned = stringToBool(record[6])
		l.Offered = stringToBool(record[7])
		l.Registered = stringToBool(record[8])
		l.Used = stringToBool(record[9])

		mac, _ := net.ParseMAC(record[1])
		l.MAC = mac

		start, _ := time.ParseInLocation(time.RFC3339, record[3], time.Local)
		l.Start = start

		end, _ := time.ParseInLocation(time.RFC3339, record[4], time.Local)
		l.End = end

		leases = append(leases, l)
	}

	return leases, nil
}

func stringToBool(s string) bool {
	if s == "y" {
		return true
	}
	return false
}
