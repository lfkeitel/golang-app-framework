package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/lfkeitel/golang-app-framework/src/db"
	"github.com/lfkeitel/golang-app-framework/src/server"
	"github.com/lfkeitel/golang-app-framework/src/utils"
	"github.com/lfkeitel/verbose"
)

var (
	configFile string
	dev        bool
	verFlag    bool
	testConfig bool

	version   = ""
	buildTime = ""
	builder   = ""
	goversion = ""
	appName   = ""
)

func init() {
	flag.StringVar(&configFile, "c", "", "Configuration file path")
	flag.BoolVar(&dev, "d", false, "Run in development mode")
	flag.BoolVar(&verFlag, "version", false, "Display version information")
	flag.BoolVar(&verFlag, "v", verFlag, "Display version information")
	flag.BoolVar(&testConfig, "t", false, "Test main configuration")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Parse CLI flags
	flag.Parse()

	if verFlag {
		displayVersionInfo()
		return
	}

	if configFile == "" || !utils.FileExists(configFile) {
		configFile = utils.FindConfigFile()
	}
	if configFile == "" {
		fmt.Println("No configuration file found")
		os.Exit(1)
	}

	if testConfig {
		testMainConfig()
		return
	}

	var err error
	e := utils.NewEnvironment(utils.EnvProd)
	if dev {
		e.Env = utils.EnvDev
	}

	e.Config, err = utils.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %s\n", err)
		os.Exit(1)
	}

	e.Log = utils.NewLogger(e.Config, "app")
	utils.SystemLogger = e.Log
	e.Log.Debugf("Configuration loaded from %s", configFile)

	c := e.SubscribeShutdown()
	go func(e *utils.Environment) {
		<-c
		e.Log.Notice("Shutting down...")
		time.Sleep(2)
	}(e)

	e.DB, err = db.NewDatabaseAccessor(e)
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading database")
	}
	e.Log.WithFields(verbose.Fields{
		"type":    e.Config.Database.Type,
		"address": e.Config.Database.Address,
	}).Debug("Loaded database")

	e.Sessions, err = utils.NewSessionStore(e)
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading session store")
	}

	e.Views, err = utils.NewViews(e, "templates")
	if err != nil {
		e.Log.WithField("error", err).Fatal("Error loading frontend templates")
	}

	// Start web server
	server.NewServer(e, server.LoadRoutes(e)).Run()
}

func displayVersionInfo() {
	fmt.Printf(`%s - (C) 2016 App Author

Component:   Web Server
Version:     %s
Built:       %s
Compiled by: %s
Go version:  %s
`, appName, version, buildTime, builder, goversion)
}

func testMainConfig() {
	_, err := utils.NewConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Configuration looks good")
}
