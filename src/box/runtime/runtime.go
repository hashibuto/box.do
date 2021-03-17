package runtime

import (
	"box/config"
	"box/manifest"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"gopkg.in/yaml.v2"
)

var isInitialized = false

type runData struct {
	Project string `yaml:"project"`
}

type Runtime struct {
	Manifest *manifest.Manifest
	Client   *client.Client
	Context  context.Context
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

var cronService manifest.Service = manifest.Service{
	Image: "hashibuto/box-cron",
	Volumes: []string{
		"@/letsencrypt:/etc/letsencrypt",
		"@/www/acme:/var/www/acme",
		"/var/run/docker.sock:/var/run/docker.sock",
	},
}

// All box managed containers will start with this prefix
const boxContainerPrefix = "box__"

// New returns a new instance of the runtime structure for the supplied project
func New(mfst *manifest.Manifest, isProduction bool) (*Runtime, error) {
	// Basically we want this to behave as a singleton, which will disrupt any existing
	// runtime that's running against a different project name (one at a time only).
	if isInitialized == true {
		return nil, fmt.Errorf("Only one runtime can be initialized per execution")
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	runInfo := runData{}
	runFilename := path.Join(configDir, ".run.yml")
	data, err := ioutil.ReadFile(runFilename)
	if err == nil {
		err = yaml.Unmarshal(data, &runInfo)
		if err != nil {
			// If the YAML can't be parsed, revert back to the empty structure but ignore the error
			runInfo = runData{}
		}
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	if runInfo.Project != mfst.Project {
		oldContainerIDs := []string{}

		// The previous running project was not this project, so kill all of the management containers
		ctx := context.Background()
		containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
		if err != nil {
			return nil, err
		}

		for _, container := range containers {
			for _, name := range container.Names {
				if strings.HasPrefix(name, boxContainerPrefix) {
					oldContainerIDs = append(oldContainerIDs, container.ID)
					break
				}
			}
		}

		if len(oldContainerIDs) > 0 {
			return nil, fmt.Errorf("The project \"%v\" is still running, please stop it before running another project.\n", runInfo.Project)
		}
	}

	runInfo.Project = mfst.Project
	outBytes, err := yaml.Marshal(&runInfo)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(runFilename, outBytes, os.FileMode(0600))
	if err != nil {
		return nil, err
	}

	isInitialized = true
	return &Runtime{
		Manifest: mfst,
		Client:   cli,
		Context:  context.Background(),
	}, nil
}
