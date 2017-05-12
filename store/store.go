package store

import (
	"net"

	"github.com/boltdb/bolt"
)

var leaseBucket = []byte("leases")

type Store struct {
	db *bolt.DB
}

func NewStore(path string) (*Store, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(leaseBucket)
		return err
	})

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) GetLease(ip net.IP) (*Lease, error) {
	var data []byte

	s.db.View(func(tx *bolt.Tx) error {
		data = tx.Bucket(leaseBucket).Get([]byte(ip.To4()))
		return nil
	})

	lease := NewLease()
	if err := lease.unserialize(data); err != nil {
		return nil, err
	}
	return lease, nil
}

func (s *Store) PutLease(l *Lease) error {
	data := l.serialize()

	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(leaseBucket).Put([]byte(l.IP.To4()), data)
	})
}

func (s *Store) ForEachLease(foreach func(*Lease)) {
	s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(leaseBucket)
		bucket.ForEach(func(k []byte, v []byte) error {
			lease := NewLease()
			if err := lease.unserialize(v); err == nil {
				foreach(lease)
			}
			return nil
		})
		return nil
	})
}
