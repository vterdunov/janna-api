package config

import (
	"errors"
	"os"
)

// Config provide configuration to the application
type Config struct {
	Protocols protocols
	Debug     bool
	VMWare    resources
}

type resources struct {
	URL      string
	Insecure bool
	DC       string
	DS       string
	RP       string
	Folder   string
	Host     string
}

type protocols struct {
	HTTP    http
	JSONRPC jsonrpc
}

type http struct {
	Port string
}

type jsonrpc struct {
	Port string
}

// Load configuration
func Load() (*Config, error) {
	config := &Config{}

	debug := os.Getenv("DEBUG")
	if debug == "1" || debug == "true" {
		config.Debug = true
	}

	port := os.Getenv("PORT")
	if port != "" {
		config.Protocols.HTTP.Port = port
	}

	jsonrpcPort := os.Getenv("JSONRPC_PORT")
	if jsonrpcPort != "" {
		config.Protocols.JSONRPC.Port = jsonrpcPort
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
	vmwareDC, ok := os.LookupEnv("VMWARE_DC")
	if !ok {
		return nil, errors.New("provide 'VMWARE_DC' environment variable")
	}
	config.VMWare.DC = vmwareDC

	// VMWare Datastore
	vmwareDS, exist := os.LookupEnv("VMWARE_DS")
	if !exist {
		return nil, errors.New("provide 'VMWARE_DS' environment variable")
	}
	config.VMWare.DS = vmwareDS

	// VMWare Resource Pool
	vmwareRP, exist := os.LookupEnv("VMWARE_RP")
	if !exist {
		return nil, errors.New("provide 'VMWARE_RP' environment variable")
	}
	config.VMWare.RP = vmwareRP

	// VMWare VM Folder
	vmwareFolder, exist := os.LookupEnv("VMWARE_FOLDER")
	if !exist {
		config.VMWare.Folder = ""
	} else {
		config.VMWare.Folder = vmwareFolder
	}

	// VMWare ESXi Host
	vmwareHost, exist := os.LookupEnv("VMWARE_HOST")
	if !exist {
		config.VMWare.Host = ""
	} else {
		config.VMWare.Host = vmwareHost
	}

	return config, nil
}
