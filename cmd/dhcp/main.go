package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/packet-guardian/pg-dhcp/internal/config"
	"github.com/packet-guardian/pg-dhcp/internal/server"
	"github.com/packet-guardian/pg-dhcp/internal/utils"
	management "github.com/packet-guardian/pg-dhcp/managment"
	"github.com/packet-guardian/pg-dhcp/store"
)

var (
	configFile         string
	testMainConfigFlag bool
	testDHCPConfigFlag bool
	verFlag            bool
	cpuprofile         string

	version   = ""
	buildTime = ""
	builder   = ""
	goversion = ""
)

func init() {
	flag.StringVar(&configFile, "c", "", "Configuration file path")
	flag.BoolVar(&testMainConfigFlag, "t", false, "Test main configuration file")
	flag.BoolVar(&testDHCPConfigFlag, "td", false, "Test DHCP server configuration file")
	flag.BoolVar(&verFlag, "version", false, "Display version information")
	flag.BoolVar(&verFlag, "v", verFlag, "Display version information")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "CPU profile path")
}

func main() {
	flag.Parse()

	if verFlag {
		displayVersionInfo()
		return
	}

	if testMainConfigFlag {
		testMainConfig()
		return
	}

	if testDHCPConfigFlag {
		testDHCPConfig()
		return
	}

	if cpuprofile != "" {
		var err error
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}

		go func() {
			<-time.After(1 * time.Minute)
			pprof.StopCPUProfile()
			f.Close()
		}()
	}

	var err error
	e := config.NewEnvironment(config.EnvProd)

	if configFile == "" || !utils.FileExists(configFile) {
		configFile = config.FindConfigFile()
	}
	if configFile == "" {
		fmt.Println("No configuration file found")
		os.Exit(1)
	}

	e.Config, err = config.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %s\n", err.Error())
		os.Exit(1)
	}

	e.Log = config.NewLogger(e.Config, "dhcp")
	e.Log.WithField("path", configFile).Debug("Configuration loaded")

	if !utils.FileExists(e.Config.Server.NetworksFile) {
		e.Log.WithField("path", e.Config.Server.NetworksFile).Fatal("DHCP networks file not found")
	}

	networks, err := server.ParseFile(e.Config.Server.NetworksFile)
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading DHCP configuration")
	}

	e.Log.Info("Opening database")
	store, err := openDatabase(e.Config)
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading lease database")
	}

	e.Log.Info("Starting management server")
	management.SetLogger(e.Log)
	go func() {
		if err := management.StartRPCServer(e.Config.Management, store); err != nil {
			e.Log.WithField("error", err).Fatal("Error starting management interface")
		}
	}()

	serverConfig := &server.ServerConfig{
		Log:            e.Log,
		Store:          store,
		Env:            server.EnvProd,
		BlockBlacklist: e.Config.Server.BlockBlacklisted,
		Workers:        e.Config.Server.Workers,
	}

	handler := server.NewDHCPServer(networks, serverConfig)
	if err := handler.LoadLeases(); err != nil {
		e.Log.WithField("error", err).Fatal("Couldn't load leases")
	}

	go func(e *config.Environment) {
		<-e.SubscribeShutdown()
		e.Log.Notice("Shutting down...")
		handler.Close()
	}(e)

	if err := handler.ListenAndServe(); err != nil {
		e.Log.Fatal(err.Error())
	}
}

func displayVersionInfo() {
	fmt.Printf(`PG Dhcp - (C) 2016 The Packet Guardian Authors

Version:     %s
Built:       %s
Compiled by: %s
Go version:  %s
`, version, buildTime, builder, goversion)
}

func testMainConfig() {
	_, err := config.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Configuration looks good")
}

func testDHCPConfig() {
	_, err := server.ParseFile(configFile)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration looks good")
}

func openDatabase(cfg *config.Config) (store.Store, error) {
	switch cfg.Database.Type {
	case "boltdb":
		return store.NewBoltStore(cfg.Database.Path)
	case "memory":
		return store.NewMemoryStore()
	case "mysql":
		return openMySQLStore(cfg)
	case "pg":
		return openPGStore(cfg)
	}

	return nil, fmt.Errorf("Database type '%s' not supported", cfg.Database.Type)
}

func openMySQLStore(cfg *config.Config) (store.Store, error) {
	return store.NewMySQLStore(
		makeSQLConfig(cfg),
		cfg.Database.LeaseTable,
		cfg.Database.DeviceTable,
	)
}

func openPGStore(cfg *config.Config) (store.Store, error) {
	return store.NewPGStore(
		makeSQLConfig(cfg),
		cfg.Database.LeaseTable,
		cfg.Database.DeviceTable,
		cfg.Database.BlacklistTable,
	)
}

func makeSQLConfig(cfg *config.Config) *mysql.Config {
	netAddr := fmt.Sprintf("%s(%s:%d)", cfg.Database.Protocol, cfg.Database.Address, cfg.Database.Port)
	dsn := fmt.Sprintf("%s:%s@%s/%s?timeout=30s", cfg.Database.Username, cfg.Database.Password, netAddr, cfg.Database.Name)
	sqlCfg, _ := mysql.ParseDSN(dsn)
	return sqlCfg
}
