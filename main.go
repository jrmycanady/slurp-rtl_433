package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jrmycanady/slurp-rtl_433_v2/config"
	"github.com/jrmycanady/slurp-rtl_433_v2/device"
	"github.com/jrmycanady/slurp-rtl_433_v2/dump"
	"github.com/jrmycanady/slurp-rtl_433_v2/file"
	"github.com/jrmycanady/slurp-rtl_433_v2/logger"
	"github.com/ogier/pflag"
)

var (
	// All string flags defaults are a blank unicode character. '⠀'
	cPath             = pflag.StringP("config", "c", "⠀", "The path to he config file.")
	cDataLocation     = pflag.StringP("data-location", "d", "⠀", "The path and search string for the data to monitor.")
	cMetaDataLocation = pflag.StringP("meta-data-location", "m", "⠀", "The meta data folder location.")
	cFQDN             = pflag.StringP("fqdn", "f", "⠀", "The FQDN to the InfluxDB server.")
	cPort             = pflag.IntP("port", "P", -1, "The port to the InfluxDB server.")
	cUsername         = pflag.StringP("username", "u", "⠀", "The username used to connect to InfluxDB with.")
	cPassword         = pflag.StringP("password", "p", "⠀", "The password used to connect to InfluxDB with.")
	cDatabase         = pflag.StringP("database", "b", "⠀", "The name of the InfluxDB database.")
	cVerbose          = pflag.BoolP("verbose", "v", false, "Enable verbose logging.")
	cDebug            = pflag.BoolP("debug", "D", false, "Enable debug logging.")
)

// Usage replaces the default usage function for the flag package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
}

func main() {
	var err error

	pflag.Usage = Usage
	pflag.Parse()

	// Loading configuration from file and args.
	globalConfig, err := loadConfig()
	if err != nil {
		fmt.Printf("failed to load configuration\n")
		if *cDebug {
			fmt.Println(err)
			return
		}
	}

	// Configuring the logger to output file or stdout.
	output, err := buildLogger(globalConfig)
	if err != nil {
		fmt.Printf("failed to start logging\n")
		if *cDebug {
			fmt.Println(err)
			return
		}
	}
	defer output.Close()

	// Build signal channel to catch term signal.
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	logger.Info.Println("starting dumper")
	dumpChan := make(chan device.DataPoint)
	dumper := dump.NewDumper(globalConfig.InfluxDB, dumpChan)
	if err := dumper.StartDump(); err != nil {
		logger.Error.Printf("failed to start dumper: %s", err)
		return
	}

	f := file.NewFiler(globalConfig, dumpChan)
	if err := f.Start(); err != nil {
		logger.Error.Printf("failed to start filer: %s", err)
		return
	}

	// Waiting for term signal to gracefully shutdown.
	<-signals
	logger.Info.Println("received term signal, shutting down now")

	// Stop filer.
	logger.Info.Println("stopping filer")
	f.Stop()

	// Stop dumper.
	logger.Info.Println("stopping dumper")

	logger.Info.Println("slurp-rtl_433 going to bed, good night")

}

// buildLogger creates new loggers based on the parameters found in the current
// configuration. If this never called the default is to log all levels out
// to stdout.
func buildLogger(cfg config.Config) (*os.File, error) {
	var output *os.File
	var err error

	// Configuring to use file for logging if needed.
	if cfg.LogFilePath != "" {
		output, err = logger.ConfigureWithFile(cfg.LogFilePath, cfg.LogLevels)
		if err != nil {
			return nil, fmt.Errorf("failed to setup file %s for logging: %v", cfg.LogFilePath, err)
		}
		fmt.Printf("sending logs to %s", cfg.LogFilePath)
		return output, nil
	}
	logger.UpdateWithLevelList(os.Stdout, cfg.LogLevels)
	return output, nil
}

// loadConfig loads the configuration file located the path provided on the commandline and
// and augments based on any other commandline arguments. All commandline arguments supersede
// the value found in the configuration file.
func loadConfig() (config.Config, error) {

	// Creating an default config or loading the config from file.
	cfg := config.NewConfig()
	if *cPath != "⠀" {
		cfg, err := config.LoadConfigFromFile(*cPath)
		if err != nil {
			return cfg, err
		}
	}

	// Superseding any config options provided.
	if *cFQDN != "⠀" {
		cfg.InfluxDB.FQDN = *cFQDN
	}
	if *cPort != -1 {
		cfg.InfluxDB.Port = *cPort
	}
	if *cUsername != "⠀" {
		cfg.InfluxDB.Username = *cUsername
	}
	if *cPassword != "⠀" {
		cfg.InfluxDB.Password = *cPassword
	}
	if *cDatabase != "⠀" {
		cfg.InfluxDB.Database = *cDatabase
	}
	if *cDataLocation != "⠀" {
		name, dir, err := config.SplitLogPath(*cDataLocation)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse data directory")

		}
		cfg.DataFileDir = dir
		cfg.DataFileName = name
		cfg.DataLocation = *cDataLocation
	}
	if *cMetaDataLocation != "⠀" {
		cfg.FileMetaDataPath = *cMetaDataLocation
	}
	if *cVerbose {
		cfg.LogLevels = append(cfg.LogLevels, "verbose")
	}
	if *cDebug {
		cfg.LogLevels = append(cfg.LogLevels, "debug")
	}

	return cfg, nil
}
