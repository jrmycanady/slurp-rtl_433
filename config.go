package main

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// Config represents the configuration for a slurp-rtl_433 instance.
type Config struct {
	DataLocation string
	LogFilePath  string
	LogLevels    []string
	InfluxDB     InfluxDBConfig
	Meta         map[string]map[string]map[string]map[string]interface{}
}

// InfluxDBConfig represents the configuration for an InfluxDB connection.
type InfluxDBConfig struct {
	FQDN     string
	Port     int
	Username string
	Password string
	Database string
}

// NewConfig generates a new empty configuration.
func NewConfig() Config {
	return Config{
		DataLocation: "rtl_433.out*",
		LogLevels:    []string{"info", "error"},
		InfluxDB: InfluxDBConfig{
			FQDN:     "localhost",
			Port:     8086,
			Database: "slurp-rtl_433",
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

	return config, nil
}
