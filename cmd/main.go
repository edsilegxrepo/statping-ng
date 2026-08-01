package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/handlers"
	"github.com/statping-ng/statping-ng/notifiers"
	"github.com/statping-ng/statping-ng/source"
	"github.com/statping-ng/statping-ng/types/configs"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/metrics"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

var (
	// VERSION stores the current version of Statping
	VERSION string = "dev"
	// COMMIT stores the git commit hash for this version of Statping
	COMMIT  string
	log     = utils.Log.WithField("type", "cmd")
	confgs  *configs.DbConfig
	stopped chan bool
)

func init() {
	stopped = make(chan bool, 1)
	core.New(VERSION, COMMIT)
	utils.InitEnvs()
	utils.Params.Set("VERSION", VERSION)
	utils.Params.Set("COMMIT", COMMIT)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(assetsCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(onceCmd)
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(systemctlCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(resetAdminCmd)

	parseFlags(rootCmd)
}

// exit will return an error and return an exit code 1 due to this error
func exit(err error) {
	log.Fatalln(err)
	os.Exit(1)
}

// Close will gracefully stop all services, HTTP server, database connection, and log file
func Close() {
	services.StopAll()
	handlers.StopHTTPServer(nil)
	utils.StopLogShipper()
	utils.CloseLogs()
	confgs.Close()
	fmt.Println("Shutting down Statping")
}

// main will run the Statping application
func main() {
	Execute()
}

// start will run the Statping application server mode
func start() {
	go sigterm()
	var err error
	if err := source.Assets(); err != nil {
		exit(err)
	}

	utils.VerboseMode = verboseMode

	if err := utils.InitLogs(); err != nil {
		log.Errorf("Statping Log Error: %v\n", err)
	}

	log.Info(fmt.Sprintf("Starting Statping %s", VERSION))

	// Initialize master key for encryption (mandatory)
	if err := utils.InitMasterKey(); err != nil {
		log.Errorf("Master key initialization failed: %v", err)
		log.Error("Run 'statping init --key-file /path/to/master.key' to generate a key")
		exit(err)
	}
	defer utils.ZeroMasterKey()

	utils.Params.Set("SERVER_IP", ipAddress)
	utils.Params.Set("SERVER_PORT", port)

	confgs, err = configs.LoadConfigs(configFile)
	if err != nil {
		log.Infoln("Starting in Setup Mode")
		if err = handlers.RunHTTPServer(); err != nil {
			exit(err)
		}
	}

	if err = configs.ConnectConfigs(confgs, true); err != nil {
		exit(err)
	}

	if err = confgs.ResetCore(); err != nil {
		exit(err)
	}

	if err = confgs.DatabaseChanges(); err != nil {
		exit(err)
	}

	if err := confgs.MigrateDatabase(); err != nil {
		exit(err)
	}

	if err := mainProcess(); err != nil {
		exit(err)
	}

	<-stopped
}

// sigterm will attempt to close the database connections gracefully
func sigterm() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Infoln("Received shutdown signal")
	Close()
	stopped <- true
}

// mainProcess will initialize the Statping application and run the HTTP server
func mainProcess() error {
	if err := InitApp(); err != nil {
		return err
	}

	_, _ = services.LoadServicesYaml()

	if err := handlers.RunHTTPServer(); err != nil {
		log.Fatalln(err)
		return errors.Wrap(err, "http server")
	}
	return nil
}

// InitApp will start the Statping instance with a valid database connection
// This function will gather all services in database, add/init Notifiers,
// and start the database cleanup routine
func InitApp() error {
	// fetch Core row information about this instance.
	if _, err := core.Select(); err != nil {
		return err
	}
	// init log shipping if configured (Loki, Elasticsearch, Splunk, Cribl)
	// Environment variables take precedence, database config is fallback
	utils.InitLogShipperWithConfig(&utils.LogShipConfig{
		Enabled:    core.App.LogShipEnabled.Bool,
		Type:       core.App.LogShipType,
		Endpoint:   core.App.LogShipEndpoint,
		Token:      core.App.LogShipToken,
		Index:      core.App.LogShipIndex,
		Sourcetype: core.App.LogShipSourcetype,
		Labels:     core.App.LogShipLabels,
	})
	// init prometheus metrics
	metrics.InitMetrics()
	// connect each notifier, added them into database if needed
	notifiers.InitNotifiers()
	// select all services in database and store services in a mapping of Service pointers
	if _, err := services.SelectAllServices(true); err != nil {
		return err
	}
	// start routines for each service checking process
	services.CheckServices()
	// start routine to delete old records (failures, hits)
	go database.Maintenance()
	// start daily digest scheduler
	notifiers.StartDigestScheduler()
	core.App.Setup = true
	core.App.Started = utils.Now()
	return nil
}
