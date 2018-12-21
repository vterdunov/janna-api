package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

// Config provide configuration to the application
type Config struct {
	Protocols protocols
	Debug     bool
	DebugHTTP bool
	VMWare    resources
	TaskTTL   time.Duration
}

type resources struct {
	URL      string
	Insecure bool
	DC       string
	Folder   string
}

type protocols struct {
	HTTP http
}

type http struct {
	Port string
}

// Load configuration
func Load() (*Config, error) {
	config := &Config{}

	debug := os.Getenv("DEBUG")
	if debug == "1" || debug == "true" { //nolint:goconst
		config.Debug = true
	}

	DebugHTTP := os.Getenv("DEBUG_HTTP")
	if DebugHTTP == "1" || DebugHTTP == "true" { //nolint:goconst
		config.DebugHTTP = true
	}

	port := os.Getenv("PORT")
	if port != "" {
		config.Protocols.HTTP.Port = port
	}

	// VMWare URL
	vmwareURL, exist := os.LookupEnv("VMWARE_URL")
	if !exist {
		return nil, errors.New("provide 'VMWARE_URL' environment variable")
	}
	config.VMWare.URL = vmwareURL

	// VMWare insecure
	vmwareInsecure := os.Getenv("VMWARE_INSECURE")
	if vmwareInsecure == "1" || vmwareInsecure == "true" {
		config.VMWare.Insecure = true
	}
	config.VMWare.URL = vmwareURL

	// VMWare Datacenter
	vmwareDC, ok := os.LookupEnv("VMWARE_DATACENTER")
	if !ok {
		return nil, errors.New("provide 'VMWARE_DATACENTER' environment variable")
	}
	config.VMWare.DC = vmwareDC

	// VMWare VM Folder
	vmwareFolder, exist := os.LookupEnv("VMWARE_FOLDER")
	if exist {
		config.VMWare.Folder = vmwareFolder
	}

	// Background jobs time to live
	defaultTTL := time.Minute * 30
	taskTTL, exist := os.LookupEnv("TASKS_TTL")
	minutes, err := strconv.Atoi(taskTTL)
	if err != nil || !exist {
		config.TaskTTL = defaultTTL
	} else {
		config.TaskTTL = time.Minute * time.Duration(minutes)
	}

	return config, nil
}
