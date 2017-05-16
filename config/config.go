package config

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/naoina/toml"
	"github.com/packet-guardian/pg-dhcp/utils"
)

type Config struct {
	sourceFile string
	Logging    struct {
		Disabled bool
		Level    string
		Path     string
	}
	Leases struct {
		HistoryEnabled   bool
		DeleteWithDevice bool
		DeleteAfter      string
		DatabaseFile     string
	}
	Events struct {
		Address  string
		Types    []string
		Username string
		Password string
	}
	Verification struct {
		Address          string
		ReconnectTimeout string
	}
	Server struct {
		NetworksFile string
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

	// Leases
	c.Leases.DeleteAfter = setStringOrDefault(c.Leases.DeleteAfter, "96h")
	if _, err := time.ParseDuration(c.Leases.DeleteAfter); err != nil {
		c.Leases.DeleteAfter = "96h"
	}
	c.Leases.DatabaseFile = setStringOrDefault(c.Leases.DatabaseFile, "leases.db")

	// Verification
	c.Verification.ReconnectTimeout = setStringOrDefault(c.Verification.ReconnectTimeout, "10s")
	if _, err := time.ParseDuration(c.Verification.ReconnectTimeout); err != nil {
		c.Verification.ReconnectTimeout = "10s"
	}

	// DHCP
	c.Server.NetworksFile = setStringOrDefault(c.Server.NetworksFile, "/etc/pg-dhcp/dhcp.conf")

	return c, nil
}

// Given string s, if it is empty, return v else return s.
func setStringOrDefault(s, v string) string {
	if s == "" {
		return v
	}
	return s
}
