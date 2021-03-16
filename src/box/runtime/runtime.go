package runtime

import (
	"box/config"
	"box/manifest"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/docker/docker/client"

	"gopkg.in/yaml.v2"
)

var isInitialized = false

type runData struct {
	Project string `yaml:"project"`
}

type Runtime struct {
	Manifest *manifest.Manifest
}

var routerService manifest.Service = manifest.Service{
	Hostname: "box-router",
	Image:    "hashibuto/box-router",
	Ports: []string{
		"80:80",
		"443:443",
	},
	Volumes: []string{
		"@/letsencrypt:/etc/letsencrypt",
		"@/www/acme:/var/www/acme",
	},
}

var registryService manifest.Service = manifest.Service{
	Hostname: "box-registry",
	Image:    "hashibuto/box-registry",
	Ports: []string{
		"5000:5000",
	},
	Volumes: []string{
		"@/registry:/var/lib/registry",
	},
}

// New returns a new instance of the runtime structure for the supplied project
func New(manifestFilename string, isProduction bool) (*Runtime, error) {
	// Basically we want this to behave as a singleton, which will disrupt any existing
	// runtime that's running against a different project name (one at a time only).
	if isInitialized == true {
		return nil, fmt.Errorf("Only one runtime can be initialized per execution")
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(path.Join(configDir, ".run.yml"))
	if err != nil {
		return nil, err
	}
	runInfo := runData{}
	err = yaml.Unmarshal(data, &runInfo)
	if err != nil {
		// If the YAML can't be parsed, revert back to the empty structure but ignore the error
		runInfo = runData{}
	}

	mfst, err := manifest.NewManifest(manifestFilename)
	if err != nil {
		return nil, err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	if runInfo.Project != mfst.Project {
		// The previous running project was not this project, so kill all of the management containers

	}

	isInitialized = true
	return &Runtime{
		Manifest: mfst,
	}, nil
}
