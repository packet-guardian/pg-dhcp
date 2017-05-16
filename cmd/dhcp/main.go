package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/packet-guardian/pg-dhcp/config"
	"github.com/packet-guardian/pg-dhcp/events"
	"github.com/packet-guardian/pg-dhcp/server"
	"github.com/packet-guardian/pg-dhcp/store"
	"github.com/packet-guardian/pg-dhcp/utils"
	"github.com/packet-guardian/pg-dhcp/verification"
)

var (
	configFile         string
	testMainConfigFlag bool
	testDHCPConfigFlag bool
	verFlag            bool

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

	store, err := store.NewStore(e.Config.Leases.DatabaseFile)
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading lease database")
	}

	var verifier verification.Verifier
	if e.Config.Verification.Address == "" {
		verifier = verification.NewNullVerifier()
	} else {
		var err error
		for {
			timeout, _ := time.ParseDuration(e.Config.Verification.ReconnectTimeout)
			verifier, err = verification.NewRemoteVerifier(e.Config.Verification.Address, e.Log, timeout)
			if err == nil {
				break
			}
			e.Log.WithField("error", err).Alert("Failed connected to verification server. Trying again")
			time.Sleep(2 * time.Second)
		}
	}

	var emitter events.Emitter
	if e.Config.Events.Address == "" {
		emitter = events.NewNullEmitter()
	} else {
		endpoint, err := url.Parse(e.Config.Events.Address)
		if err != nil {
			e.Log.Fatal("Invalid event endpoint url")
		}
		emitter = events.NewHTTPEmitter(
			endpoint,
			events.StringsToEventTypes(e.Config.Events.Types),
			e.Config.Events.Username,
			e.Config.Events.Password)
	}

	serverConfig := &server.ServerConfig{
		Verification: verifier,
		Log:          e.Log,
		Store:        store,
		Events:       emitter,
		Env:          server.EnvProd,
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
