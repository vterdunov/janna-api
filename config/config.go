package config

import (
	"errors"
	"os"
)

// Config provide configuration to the application
type Config struct {
	Port   string
	Debug  bool
	Vmware vmware
}

type vmware struct {
	URL      string
	Insecure bool
	DC       string
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

	// VMWare URL
	vmwareURL, ok := os.LookupEnv("VMWARE_URL")
	if !ok {
		return nil, errors.New("provide 'VMWARE_URL' environment variable")
	}
	config.Vmware.URL = vmwareURL

	// VMWare insecure
	vmwareInsecure := os.Getenv("VMWARE_INSECURE")
	if vmwareInsecure == "1" || vmwareInsecure == "true" {
		config.Vmware.Insecure = true
	}
	config.Vmware.URL = vmwareURL

	// VMWare URL
	vmwareDC, ok := os.LookupEnv("VMWARE_DC")
	if !ok {
		return nil, errors.New("provide 'VMWARE_DC' environment variable")
	}
	config.Vmware.DC = vmwareDC

	return config, nil
}
