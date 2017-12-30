package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime/pprof"
	"time"

	"github.com/packet-guardian/pg-dhcp/managment"

	"github.com/packet-guardian/pg-dhcp/internal/config"
	"github.com/packet-guardian/pg-dhcp/internal/server"
	"github.com/packet-guardian/pg-dhcp/internal/utils"
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
	e.Log.Debugf("Configuration loaded from %s", configFile)

	if !utils.FileExists(e.Config.Server.NetworksFile) {
		e.Log.Fatalf("DHCP networks file not found: %s", e.Config.Server.NetworksFile)
	}

	networks, err := server.ParseFile(e.Config.Server.NetworksFile)
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading DHCP configuration")
	}

	e.Log.Info("Opening database")
	store, err := store.NewBoltStore(e.Config.Leases.DatabaseFile)
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading lease database")
	}

	e.Log.Info("Starting management server")
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", e.Config.Management.Address, e.Config.Management.Port))
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error starting management interface")
	}
	go management.StartRPCServer(l, store)
	e.Log.Infof("Management server listening on %s:%d", e.Config.Management.Address, e.Config.Management.Port)

	serverConfig := &server.ServerConfig{
		Log:            e.Log,
		Store:          store,
		Env:            server.EnvProd,
		BlockBlacklist: e.Config.Server.BlockBlacklisted,
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
		e.Log.Fatal(err)
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
