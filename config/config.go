package config

import (
	"errors"
	"os"
)

// Config provide configuration to the application
type Config struct {
	Port   string
	Vmware vmware
	Debug  bool
}

type vmware struct {
	URI      string
	Insecure bool
}

// Load configuration
func Load() (*Config, error) {
	config := &Config{}

	debug := os.Getenv("DEBUG")
	if debug == "1" || debug == "true" {
		config.Debug = true
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		return nil, errors.New("provide 'PORT' environment variable")
	}
	config.Port = port

	vmwareURI, ok := os.LookupEnv("VMWARE_URI")
	if !ok {
		return nil, errors.New("provide 'VMWARE_URI' environment variable")
	}
	config.Vmware.URI = vmwareURI

	return config, nil
}
