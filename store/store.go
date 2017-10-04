package store

import (
	"container/list"
	"net"
	"sync"
	"time"

	bolt "github.com/coreos/bbolt"
)

var (
	leaseBucket  = []byte("leases")
	deviceBucket = []byte("devices")

	flushInterval = 500 * time.Millisecond
)

type Store struct {
	m           sync.Mutex
	db          *bolt.DB
	leaseQueue  *list.List
	deviceQueue *list.List
	done        chan struct{}
}

func NewStore(path string) (*Store, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(leaseBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(deviceBucket)
		return err
	})

	s := &Store{
		db:          db,
		leaseQueue:  list.New(),
		deviceQueue: list.New(),
		done:        make(chan struct{}),
	}
	go s.startFlushTimer()

	return s, nil
}

func (s *Store) startFlushTimer() {
	t := time.NewTimer(flushInterval)
	for {
		select {
		case <-t.C:
			s.Flush()
			t.Reset(flushInterval)
		case <-s.done:
			t.Stop()
			s.Flush()
			close(s.done)
			return
		}
	}
}

func (s *Store) Flush() {
	s.m.Lock()
	leaseQueueLen := s.leaseQueue.Len()
	if leaseQueueLen > 0 {
		leaseBatch := make([]queueItem, leaseQueueLen)
		for i := 0; i < leaseQueueLen; i++ {
			elem := s.leaseQueue.Front()
			leaseBatch[i] = elem.Value.(queueItem)
			s.leaseQueue.Remove(elem)
		}

		s.db.Batch(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(leaseBucket)
			for _, item := range leaseBatch {
				if err := bucket.Put(item.key, item.val); err != nil {
					return err
				}
			}
			return nil
		})
	}

	deviceQueueLen := s.deviceQueue.Len()
	if deviceQueueLen > 0 {
		deviceBatch := make([]queueItem, deviceQueueLen)
		for i := 0; i < deviceQueueLen; i++ {
			elem := s.deviceQueue.Front()
			deviceBatch[i] = elem.Value.(queueItem)
			s.deviceQueue.Remove(elem)
		}

		s.db.Batch(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(deviceBucket)
			for _, item := range deviceBatch {
				if err := bucket.Put(item.key, item.val); err != nil {
					return err
				}
			}
			return nil
		})
	}
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

func (s *Store) GetDevice(mac net.HardwareAddr) *Device {
	var data []byte

	s.db.View(func(tx *bolt.Tx) error {
		data = tx.Bucket(deviceBucket).Get([]byte(mac))
		return nil
	})

	if data == nil { // Device doesn't exist, set everything to false
		data = []byte{0}
	}

	device := &Device{}
	device.MAC = mac
	device.Registered = byteToBool((data[0] & 2) >> 1)
	device.Blacklisted = byteToBool(data[0] & 1)
	return device
}

type queueItem struct {
	key, val []byte
}

func (s *Store) PutLease(l *Lease) error {
	data := l.serialize()
	s.m.Lock()
	s.leaseQueue.PushBack(queueItem{[]byte(l.IP.To4()), data})
	s.m.Unlock()
	return nil
}

func (s *Store) PutDevice(d *Device) {
	state := byte(0) | (boolToByte(d.Registered) << 1) | (boolToByte(d.Blacklisted))
	item := queueItem{
		key: []byte(d.MAC),
		val: []byte{state},
	}
	s.m.Lock()
	s.deviceQueue.PushBack(item)
	s.m.Unlock()
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

func (s *Store) ForEachDevice(foreach func(*Device)) {
	s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(deviceBucket)
		bucket.ForEach(func(k []byte, v []byte) error {
			device := &Device{
				MAC:         net.HardwareAddr(k),
				Registered:  byteToBool((v[0] & 2) >> 1),
				Blacklisted: byteToBool(v[0] & 1),
			}
			foreach(device)
			return nil
		})
		return nil
	})
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func byteToBool(b byte) bool {
	return b == 1
}
