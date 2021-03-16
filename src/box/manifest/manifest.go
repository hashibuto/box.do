package manifest

import (
	"fmt"
	"io/ioutil"
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

type Service struct {
	Hostname    string            `yaml:"hostname"`
	Routing     Routing           `yaml:"routing"`
	Environment map[string]string `yaml:"environment"`
	Volumes     []string          `yaml:"volumes"`
	Image       string            `yaml:"image"`
	DependsOn   string            `yaml:"depends_on"`
	Ports       []string          `yaml:"ports"`
}

type Manifest struct {
	Project  string             `yaml:"project"`
	Services map[string]Service `yaml:"services"`
}

var hostnameRe *regexp.Regexp = regexp.MustCompile("([a-z]+){3-20}")

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

// NewManifest loads, validates, and returns a pointer to the Manifest structure.  Any failure in
// loading, parsing, or validating will result in an error.
func NewManifest(filename string) (*Manifest, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
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
	}

	return &mfst, nil
}
