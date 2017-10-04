package store

import (
	"container/list"
	"net"
	"sync"
	"time"

	bolt "github.com/coreos/bbolt"
)

var leaseBucket = []byte("leases")

type Store struct {
	m     sync.Mutex
	db    *bolt.DB
	queue *list.List
	done  chan struct{}
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

	s := &Store{
		db:    db,
		queue: list.New(),
		done:  make(chan struct{}),
	}
	go s.flush()

	return s, nil
}

func (s *Store) flush() {
	t := time.NewTimer(500 * time.Millisecond)
	for {
		select {
		case <-t.C:
			s.doFlush()
			t.Reset(500 * time.Millisecond)
		case <-s.done:
			t.Stop()
			s.doFlush()
			close(s.done)
			return
		}
	}
}

func (s *Store) doFlush() {
	s.m.Lock()
	queueLen := s.queue.Len()
	if queueLen == 0 {
		s.m.Unlock()
		return
	}

	batch := make([]queueItem, queueLen)

	for i := 0; i < queueLen; i++ {
		elem := s.queue.Front()
		batch[i] = elem.Value.(queueItem)
		s.queue.Remove(elem)
	}

	s.db.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(leaseBucket)
		for _, item := range batch {
			if err := bucket.Put(item.key, item.val); err != nil {
				return err
			}
		}
		return nil
	})
	s.m.Unlock()
}

func (s *Store) Close() error {
	s.done <- struct{}{}
	<-s.done
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

type queueItem struct {
	key, val []byte
}

func (s *Store) PutLease(l *Lease) error {
	data := l.serialize()
	s.m.Lock()
	s.queue.PushBack(queueItem{[]byte(l.IP.To4()), data})
	s.m.Unlock()
	return nil

	// return s.db.Batch(func(tx *bolt.Tx) error {
	// 	return tx.Bucket(leaseBucket).Put([]byte(l.IP.To4()), data)
	// })
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
