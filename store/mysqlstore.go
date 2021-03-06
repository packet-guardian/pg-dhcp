/* Required schema:

CREATE TABLE "device" (
	"mac" VARCHAR(17) NOT NULL UNIQUE KEY,
	"registered" TINYINT DEFAULT 0,
	"blacklisted" TINYINT DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8

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
	"time"

	"github.com/packet-guardian/pg-dhcp/models"

	"github.com/go-sql-driver/mysql"
)

type MySQLStore struct {
	db                      *sql.DB
	leaseTable, deviceTable string

	prepared     bool
	prepareMutex sync.Mutex

	getLeaseStmt      *sql.Stmt
	getAllLeasesStmt  *sql.Stmt
	putLeaseStmt      *sql.Stmt
	getDeviceStmt     *sql.Stmt
	getAllDevicesStmt *sql.Stmt
	putDeviceStmt     *sql.Stmt
	deleteDeviceStmt  *sql.Stmt
}

func NewMySQLStore(cfg *mysql.Config, leaseTable, deviceTable string) (*MySQLStore, error) {
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	s := &MySQLStore{
		db:          db,
		leaseTable:  leaseTable,
		deviceTable: deviceTable,
	}
	return s, nil
}

func (s *MySQLStore) Close() error {
	return s.db.Close()
}

func (s *MySQLStore) prepare() error {
	if s.prepared {
		return nil
	}

	s.prepareMutex.Lock()
	defer s.prepareMutex.Unlock()

	if err := s.prepareLeaseStmts(); err != nil {
		return err
	}

	if err := s.prepareDeviceStmts(); err != nil {
		return err
	}

	s.prepared = true
	return nil
}

func (s *MySQLStore) prepareLeaseStmts() error {
	var err error
	s.getLeaseStmt, err = s.db.Prepare(fmt.Sprintf(`SELECT "mac", "network", "start", "end", "hostname", "abandoned", "registered" FROM "%s" WHERE "ip" = ?`, s.leaseTable))
	if err != nil {
		return err
	}

	s.getAllLeasesStmt, err = s.db.Prepare(fmt.Sprintf(`SELECT "ip", "mac", "network", "start", "end", "hostname", "abandoned", "registered" FROM "%s"`, s.leaseTable))
	if err != nil {
		return err
	}

	s.putLeaseStmt, err = s.db.Prepare(fmt.Sprintf(
		`INSERT INTO "%s" (ip, mac, network, start, end, hostname, abandoned, registered)
			VALUES (?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY
			UPDATE mac=VALUES(mac), network=VALUES(network), start=VALUES(start), end=VALUES(end), hostname=VALUES(hostname), abandoned=VALUES(abandoned), registered=VALUES(registered)`, s.leaseTable))
	if err != nil {
		return err
	}

	return nil
}

func (s *MySQLStore) prepareDeviceStmts() error {
	var err error
	s.getDeviceStmt, err = s.db.Prepare(fmt.Sprintf(`SELECT "registered", "blacklisted" FROM "%s" WHERE "mac" = ?`, s.deviceTable))
	if err != nil {
		return err
	}

	s.getAllDevicesStmt, err = s.db.Prepare(fmt.Sprintf(`SELECT "mac", "registered", "blacklisted" FROM "%s"`, s.deviceTable))
	if err != nil {
		return err
	}

	s.putDeviceStmt, err = s.db.Prepare(fmt.Sprintf(
		`INSERT INTO "%s" ("mac", "registered", "blacklisted") VALUES (?,?,?)
		ON DUPLICATE KEY UPDATE registered=VALUES(registered), blacklisted=VALUES(blacklisted)`, s.deviceTable))
	if err != nil {
		return err
	}

	s.deleteDeviceStmt, err = s.db.Prepare(fmt.Sprintf(`DELETE FROM "%s" WHERE "mac" = ?`, s.deviceTable))
	if err != nil {
		return err
	}

	return nil
}

func (s *MySQLStore) GetLease(ip net.IP) (*models.Lease, error) {
	if err := s.prepare(); err != nil {
		return nil, err
	}

	row := s.getLeaseStmt.QueryRow(ip.String())
	var (
		macStr      string
		network     string
		start       int64
		end         int64
		hostname    string
		isAbandoned bool
		registered  bool
	)

	err := row.Scan(
		&macStr,
		&network,
		&start,
		&end,
		&hostname,
		&isAbandoned,
		&registered,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	mac, _ := net.ParseMAC(macStr)

	lease := models.NewLease()
	lease.IP = ip
	lease.MAC = mac
	lease.Network = network
	lease.Start = time.Unix(start, 0)
	lease.End = time.Unix(end, 0)
	lease.Hostname = hostname
	lease.IsAbandoned = isAbandoned
	lease.Registered = registered
	return lease, nil
}

func (s *MySQLStore) PutLease(l *models.Lease) error {
	if err := s.prepare(); err != nil {
		return err
	}

	_, err := s.putLeaseStmt.Exec(
		l.IP.String(),
		l.MAC.String(),
		l.Network,
		l.Start.Unix(),
		l.End.Unix(),
		l.Hostname,
		l.IsAbandoned,
		l.Registered,
	)
	return err
}

func (s *MySQLStore) ForEachLease(foreach func(*models.Lease)) error {
	if err := s.prepare(); err != nil {
		return err
	}

	rows, err := s.getAllLeasesStmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			ip          string
			macStr      string
			network     string
			start       int64
			end         int64
			hostname    string
			isAbandoned bool
			registered  bool
		)

		err := rows.Scan(
			&ip,
			&macStr,
			&network,
			&start,
			&end,
			&hostname,
			&isAbandoned,
			&registered,
		)
		if err != nil {
			return err
		}

		mac, _ := net.ParseMAC(macStr)

		lease := models.NewLease()
		lease.IP = net.ParseIP(ip)
		lease.MAC = mac
		lease.Network = network
		lease.Start = time.Unix(start, 0)
		lease.End = time.Unix(end, 0)
		lease.Hostname = hostname
		lease.IsAbandoned = isAbandoned
		lease.Registered = registered
		foreach(lease)
	}

	return nil
}

func (s *MySQLStore) GetDevice(mac net.HardwareAddr) (*models.Device, error) {
	if err := s.prepare(); err != nil {
		return nil, err
	}

	row := s.getDeviceStmt.QueryRow(mac.String())
	var (
		registered  bool
		blacklisted bool
	)

	err := row.Scan(
		&registered,
		&blacklisted,
	)
	if err == sql.ErrNoRows {
		err = nil
	}

	device := &models.Device{}
	device.MAC = mac
	device.Registered = registered
	device.Blacklisted = blacklisted
	return device, err
}

func (s *MySQLStore) PutDevice(d *models.Device) error {
	if err := s.prepare(); err != nil {
		return err
	}

	_, err := s.putDeviceStmt.Exec(
		d.MAC.String(),
		d.Registered,
		d.Blacklisted,
	)
	return err
}

func (s *MySQLStore) DeleteDevice(d *models.Device) error {
	if err := s.prepare(); err != nil {
		return err
	}

	_, err := s.deleteDeviceStmt.Exec(d.MAC.String())
	return err
}

func (s *MySQLStore) ForEachDevice(foreach func(*models.Device)) error {
	if err := s.prepare(); err != nil {
		return err
	}

	rows, err := s.getAllDevicesStmt.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			macStr      string
			registered  bool
			blacklisted bool
		)

		err := rows.Scan(
			&macStr,
			&registered,
			&blacklisted,
		)
		if err != nil {
			return err
		}

		mac, _ := net.ParseMAC(macStr)

		device := &models.Device{}
		device.MAC = mac
		device.Registered = registered
		device.Blacklisted = blacklisted
		foreach(device)
	}
	return nil
}
