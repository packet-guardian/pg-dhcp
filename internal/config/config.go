package config

import (
	"errors"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"github.com/naoina/toml"
	"github.com/packet-guardian/pg-dhcp/internal/utils"
)

type Config struct {
	sourceFile string
	Logging    struct {
		Disabled bool
		Level    string
		Path     string
	}
	Database struct {
		Type     string
		Path     string
		Username string
		Password string
		Protocol string
		Address  string
		Port     int
		Name     string

		LeaseTable     string
		DeviceTable    string
		BlacklistTable string
	}
	Leases struct {
		DeleteAfter string // TODO: Run a job to clean up old leases
	}
	Server struct {
		BlockBlacklisted bool
		NetworksFile     string
		Workers          int
	}
	Management struct {
		Address    string
		Port       int
		AllowedIPs []string // TODO: Implement
	}
}

func FindConfigFile() string {
	if os.Getenv("PG_DHCP_CONFIG") != "" && utils.FileExists(os.Getenv("PG_DHCP_CONFIG")) {
		return os.Getenv("PG_DHCP_CONFIG")
	} else if utils.FileExists("./config.toml") {
		return "./config.toml"
	} else if utils.FileExists("/etc/pg-dhcp/config.toml") {
		return "/etc/pg-dhcp/config.toml"
	}
	return ""
}

func NewEmptyConfig() *Config {
	return &Config{}
}

func NewConfig(configFile string) (conf *Config, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()

	if configFile == "" {
		configFile = "config.toml"
	}

	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var con Config
	if err := toml.Unmarshal(buf, &con); err != nil {
		return nil, err
	}
	con.sourceFile = configFile
	return setSensibleDefaults(&con)
}

func setSensibleDefaults(c *Config) (*Config, error) {
	// Anything not set here implies its zero value is the default

	// Logging
	c.Logging.Level = setStringOrDefault(c.Logging.Level, "notice")
	c.Logging.Path = setStringOrDefault(c.Logging.Path, "logs/pg.log")

	// Database
	c.Database.Type = setStringOrDefault(c.Database.Type, "boltdb")
	c.Database.Path = setStringOrDefault(c.Database.Path, "database.db")
	c.Database.Username = setStringOrDefault(c.Database.Username, "root")
	c.Database.Password = setStringOrDefault(c.Database.Password, "password")
	c.Database.Protocol = setStringOrDefault(c.Database.Protocol, "tcp")
	c.Database.Address = setStringOrDefault(c.Database.Address, "localhost")
	c.Database.Port = setIntOrDefault(c.Database.Port, 3306)
	c.Database.Name = setStringOrDefault(c.Database.Name, "pg")

	c.Database.LeaseTable = setStringOrDefault(c.Database.LeaseTable, "lease")
	c.Database.DeviceTable = setStringOrDefault(c.Database.DeviceTable, "device")
	c.Database.BlacklistTable = setStringOrDefault(c.Database.BlacklistTable, "blacklist")

	// Leases
	c.Leases.DeleteAfter = setStringOrDefault(c.Leases.DeleteAfter, "96h")
	if _, err := time.ParseDuration(c.Leases.DeleteAfter); err != nil {
		c.Leases.DeleteAfter = "96h"
	}

	// DHCP
	c.Server.NetworksFile = setStringOrDefault(c.Server.NetworksFile, "/etc/pg-dhcp/dhcp.conf")
	c.Server.Workers = setIntOrDefault(c.Server.Workers, runtime.GOMAXPROCS(0))

	// Management
	c.Management.Address = setStringOrDefault(c.Management.Address, "0.0.0.0")
	c.Management.Port = setIntOrDefault(c.Management.Port, 8677)

	return c, nil
}

// Given string s, if it is empty, return v else return s.
func setStringOrDefault(s, v string) string {
	if s == "" {
		return v
	}
	return s
}

// Given int s, if it is zero, return v else return s.
func setIntOrDefault(s, v int) int {
	if s == 0 {
		return v
	}
	return s
}
