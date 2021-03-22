package manifest

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"gopkg.in/yaml.v2"
)

type Build struct {
	Dockerfile string `yaml:"dockerfile"`
	Context    string `yaml:"context"`
}

type Path struct {
	Pattern string `yaml:"pattern"`
	Type    string `yaml:"type"`
}

type Routing struct {
	Path Path `yaml:"path"`
	Port int  `yaml:"port"`
}

type Manifest struct {
	Project    string             `yaml:"project"`
	Services   map[string]Service `yaml:"services"`
	RuntimeEnv string             `yaml:"runtime_env"`
}

var hostnameRe *regexp.Regexp = regexp.MustCompile("^([a-z]+){3,20}$")
var portsMappingRe *regexp.Regexp = regexp.MustCompile("^\\d+:\\d+$")

// Trivial rejector for non bind mount style volumes
var volumeMappingRe *regexp.Regexp = regexp.MustCompile("^[/.@][^:]*:/[^:]*$")

func validateHostname(hostname string) error {
	if hostname == "" {
		return nil
	}

	if !hostnameRe.Match([]byte(hostname)) {
		return fmt.Errorf(
			"Hostname \"%v\" must contain only lowercase alpha characters, and be between 3 and 20 characters long",
			hostname,
		)
	}

	return nil
}

func validatePort(port string) error {
	if !portsMappingRe.Match([]byte(port)) {
		return fmt.Errorf(
			"Port mapping %v is incorrect, mappings must be in the form of <host_port>:<container_port>",
			port,
		)
	}

	return nil
}

func validateVolume(volume string) error {
	if !volumeMappingRe.Match([]byte(volume)) {
		return fmt.Errorf("Volume \"%v\" is invalid.  Only bind mount volumes are supported.  Eg: /var/log/mylogs:/var/log/something", volume)
	}

	return nil
}

// NewManifest loads, validates, and returns a pointer to the Manifest structure.  Any failure in
// loading, parsing, or validating will result in an error.
func NewManifest(filename string) (*Manifest, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return nil, fmt.Errorf("Unable to locate manifest at %v", filename)
		}
		return nil, err
	}

	mfst := Manifest{}
	err = yaml.Unmarshal(data, &mfst)
	if err != nil {
		return nil, fmt.Errorf("Unable to process YAML in %v", filename)
	}

	// Validate services
	for serviceName, service := range mfst.Services {
		// Validate hostname
		if err := validateHostname(service.Hostname); err != nil {
			return nil, err
		}

		if service.Hostname != "" && (service.Routing.Path.Pattern != "" ||
			service.Routing.Port != 0) {
			return nil, fmt.Errorf("Service %v - Cannot specify a hostname and a routing configuration, since routings rely on dynamically assigned hostnames", serviceName)
		}

		for _, port := range service.Ports {
			if err = validatePort(port); err != nil {
				return nil, err
			}
		}

		for _, volume := range service.Volumes {
			if err = validateVolume(volume); err != nil {
				return nil, err
			}
		}

		service.Name = serviceName
	}

	return &mfst, nil
}
