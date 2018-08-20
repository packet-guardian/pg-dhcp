package store

import (
	"container/list"
	"net"
	"sync"
	"time"

	"github.com/packet-guardian/pg-dhcp/models"

	bolt "github.com/coreos/bbolt"
)

var (
	leaseBucket  = []byte("leases")
	deviceBucket = []byte("devices")

	flushInterval = 500 * time.Millisecond
)

type BoltStore struct {
	m          sync.Mutex
	db         *bolt.DB
	leaseQueue *list.List
	done       chan struct{}
}

type queueItem struct {
	key, val []byte
}

func NewBoltStore(path string) (*BoltStore, error) {
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

	s := &BoltStore{
		db:         db,
		leaseQueue: list.New(),
		done:       make(chan struct{}),
	}
	go s.startFlushTimer()

	return s, nil
}

func (s *BoltStore) startFlushTimer() {
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

func (s *BoltStore) Flush() {
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
	s.m.Unlock()
}

func (s *BoltStore) Close() error {
	s.done <- struct{}{}
	<-s.done
	return s.db.Close()
}

func (s *BoltStore) GetLease(ip net.IP) (*models.Lease, error) {
	var data []byte

	s.db.View(func(tx *bolt.Tx) error {
		data = tx.Bucket(leaseBucket).Get([]byte(ip.To4()))
		return nil
	})

	if data == nil {
		return nil, nil
	}

	lease := models.NewLease()
	if err := lease.Unserialize(data); err != nil {
		return nil, err
	}
	return lease, nil
}

func (s *BoltStore) PutLease(l *models.Lease) error {
	data := l.Serialize()
	s.m.Lock()
	s.leaseQueue.PushBack(queueItem{[]byte(l.IP.To4()), data})
	s.m.Unlock()
	return nil
}

func (s *BoltStore) ForEachLease(foreach func(*models.Lease)) error {
	return s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(leaseBucket)
		bucket.ForEach(func(k []byte, v []byte) error {
			lease := models.NewLease()
			if err := lease.Unserialize(v); err == nil {
				foreach(lease)
			}
			return nil
		})
		return nil
	})
}

func (s *BoltStore) GetDevice(mac net.HardwareAddr) (*models.Device, error) {
	var data []byte

	s.db.View(func(tx *bolt.Tx) error {
		data = tx.Bucket(deviceBucket).Get([]byte(mac))
		return nil
	})

	if data == nil { // Device doesn't exist, set everything to false
		data = []byte{0}
	}

	device := &models.Device{}
	device.MAC = mac
	device.Registered = byteToBool((data[0] & 2) >> 1)
	device.Blacklisted = byteToBool(data[0] & 1)
	device.LastSeen = time.Unix(0, 0)
	return device, nil
}

func (s *BoltStore) PutDevice(d *models.Device) error {
	state := byte(0) | (boolToByte(d.Registered) << 1) | (boolToByte(d.Blacklisted))
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(deviceBucket).Put([]byte(d.MAC), []byte{state})
	})
}

func (s *BoltStore) DeleteDevice(d *models.Device) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(deviceBucket).Delete([]byte(d.MAC))
	})
}

func (s *BoltStore) ForEachDevice(foreach func(*models.Device)) error {
	return s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(deviceBucket)
		bucket.ForEach(func(k []byte, v []byte) error {
			device := &models.Device{
				MAC:         net.HardwareAddr(k),
				Registered:  byteToBool((v[0] & 2) >> 1),
				Blacklisted: byteToBool(v[0] & 1),
				LastSeen:    time.Unix(0, 0),
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
