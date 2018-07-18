// Package config provides configuration parsing and information to all of
// slurp.
//
// GlobalConfig is available for sharing a single global configuration between
// all packages.
package config

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// The byte value for a forward slash.
const (
	slash byte = 47
)

var (
	GlobalConfig Config
)

// Config represents the configuration for a slurp-rtl_433 instance.
type Config struct {
	DataLocation                string
	DataFileName                string
	DataFileDir                 string
	LogFilePath                 string
	FileMetaDataPath            string
	FilerShutdownMaxWaitSeconds int
	LogFileCheckTimeSeconds     int
	LogLevels                   []string
	InfluxDB                    InfluxDBConfig
	SlurpSleepTimeSeconds       int
	Meta                        map[string]map[string]MetaDataFieldSet
}

// MetaDataFieldSet contains the set of comaprison values and new fields
// for processing on a new
type MetaDataFieldSet struct {
	CompEqualTags map[string]string
	Tags          map[string]string
}

// InfluxDBConfig represents the configuration for an InfluxDB connection.
type InfluxDBConfig struct {
	FQDN                string
	Port                int
	Username            string
	Password            string
	Database            string
	HTTPS               bool
	FlushDataPointCount int
	FlushTimeTrigger    int
}

// NewConfig generates a new empty configuration.
func NewConfig() Config {
	return Config{
		DataLocation:                "rtl_433.log",
		FileMetaDataPath:            "./meta/",
		LogLevels:                   []string{"info", "error"},
		SlurpSleepTimeSeconds:       5,
		LogFileCheckTimeSeconds:     30,
		FilerShutdownMaxWaitSeconds: 20,
		InfluxDB: InfluxDBConfig{
			FQDN:                "localhost",
			Port:                8086,
			Database:            "slurp-rtl_433",
			FlushDataPointCount: 200,
			FlushTimeTrigger:    10,
		},
	}
}

// LoadConfigFromFile loads a configuration file located at path.
func LoadConfigFromFile(path string) (Config, error) {
	config := NewConfig()

	rawConfig, err := ioutil.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %s", err)
	}

	_, err = toml.Decode(string(rawConfig), &config)
	if err != nil {
		return config, fmt.Errorf("failed to decode config file: %s", err)
	}

	// Parsing the data directory
	name, dir, err := SplitLogPath(config.DataLocation)
	if err != nil {
		return config, fmt.Errorf("failed to load data location: %v", err)
	}
	config.DataFileDir = dir
	config.DataFileName = name

	return config, nil
}

// SplitLogPath splits the path into the filename and filepath.
// If the file name is empty an error is returned.
func SplitLogPath(path string) (string, string, error) {
	var fileDirectory string
	var fileName string

	// Looking for a directory forward slash starting at the rear
	// of the string. Upon finding it mark that a directory path was
	// provided and denote the location of the directory slash.
	var j int
	var slashFound bool
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == slash {
			j = i
			slashFound = true
			break
		}
	}

	if slashFound {
		fileDirectory = path[:j+1]
		fileName = path[j+1:]
	} else {
		fileName = path
	}

	if fileName == "" {
		return fileName, fileDirectory, fmt.Errorf("no filename provided")
	}

	return fileName, fileDirectory, nil
}
