/* Expected schema (should be handled by Packet Guardian Managment application):

CREATE TABLE "device" (
	"id" INTEGER PRIMARY KEY,
	"mac" VARCHAR(17) NOT NULL UNIQUE KEY
) ENGINE=InnoDB DEFAULT CHARSET=utf8

CREATE TABLE "blacklist" (
	"id" INTEGER PRIMARY KEY AUTO_INCREMENT NOT NULL,
	"value" VARCHAR(255) NOT NULL UNIQUE KEY
) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1

CREATE TABLE "lease" (
	"ip" VARCHAR(15) NOT NULL UNIQUE KEY,
	"mac" VARCHAR(17) NOT NULL,
	"network" TEXT NOT NULL,
	"start" INTEGER NOT NULL,
	"end" INTEGER NOT NULL,
	"hostname" TEXT NOT NULL,
	"abandoned" TINYINT DEFAULT 0,
	"registered" TINYINT DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8
*/

package store

import (
	"database/sql"
	"fmt"
	"net"
	"sync"

	"github.com/packet-guardian/pg-dhcp/models"

	"github.com/go-sql-driver/mysql"
)

type PGStore struct {
	*MySQLStore
	blacklistTable string

	prepareLock     sync.Mutex
	prepared        bool
	pgGetDeviceStmt *sql.Stmt
	pgBlacklistStmt *sql.Stmt
}

func NewPGStore(cfg *mysql.Config, leaseTable, deviceTable, blacklistTable string) (*PGStore, error) {
	sqlStore, err := NewMySQLStore(cfg, leaseTable, deviceTable)
	if err != nil {
		return nil, err
	}

	s := &PGStore{
		MySQLStore:     sqlStore,
		blacklistTable: blacklistTable,
	}
	return s, nil
}

func (s *PGStore) prepare() error {
	if s.prepared {
		return nil
	}

	s.prepareLock.Lock()
	defer s.prepareLock.Unlock()

	if err := s.MySQLStore.prepareLeaseStmts(); err != nil {
		return err
	}
	s.MySQLStore.prepared = true

	var err error
	s.pgGetDeviceStmt, err = s.db.Prepare(fmt.Sprintf(`SELECT "id" FROM "%s" WHERE "mac" = ?`, s.deviceTable))
	if err != nil {
		return err
	}

	s.pgBlacklistStmt, err = s.db.Prepare(fmt.Sprintf(`SELECT "id" FROM "%s" WHERE "value" = ?`, s.blacklistTable))
	if err != nil {
		return err
	}

	s.prepared = true
	return nil
}

func (s *PGStore) GetLease(ip net.IP) (*models.Lease, error) {
	if err := s.prepare(); err != nil {
		return nil, err
	}
	return s.MySQLStore.GetLease(ip)
}
func (s *PGStore) PutLease(l *models.Lease) error {
	if err := s.prepare(); err != nil {
		return err
	}
	return s.MySQLStore.PutLease(l)
}
func (s *PGStore) ForEachLease(foreach func(*models.Lease)) error {
	if err := s.prepare(); err != nil {
		return err
	}
	return s.MySQLStore.ForEachLease(foreach)
}

func (s *PGStore) GetDevice(mac net.HardwareAddr) (*models.Device, error) {
	if err := s.prepare(); err != nil {
		return nil, err
	}

	row := s.pgGetDeviceStmt.QueryRow(mac.String())

	var id int
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		err = nil
	}

	device := &models.Device{}
	device.MAC = mac
	device.Registered = id > 0
	device.Blacklisted = s.deviceBlacklisted(mac)
	return device, err
}

func (s *PGStore) deviceBlacklisted(mac net.HardwareAddr) bool {
	var id int
	row := s.pgBlacklistStmt.QueryRow(mac.String())
	row.Scan(&id)
	return id > 0
}

func (s *PGStore) PutDevice(d *models.Device) error {
	return nil // We don't manage the devices, the management application does.
}

func (s *PGStore) DeleteDevice(d *models.Device) error {
	return nil // We don't manage the devices, the management application does.
}

func (s *PGStore) ForEachDevice(foreach func(*models.Device)) error {
	return nil
}
