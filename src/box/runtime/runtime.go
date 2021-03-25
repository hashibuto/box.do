package runtime

import (
	"box/cmd"
	"box/config"
	"box/manifest"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/sync/errgroup"

	"gopkg.in/yaml.v2"
)

var isInitialized = false

type runData struct {
	Project string `yaml:"project"`
}

type Runtime struct {
	Manifest   *manifest.Manifest
	Client     *client.Client
	Context    context.Context
	Production bool
	Config     *config.Config
}

var routerService manifest.Service = manifest.Service{
	Name:     "router",
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
	Name:     "registry",
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
	Name:  "cron",
	Image: "hashibuto/box-cron",
	Volumes: []string{
		"@/letsencrypt:/etc/letsencrypt",
		"@/www/acme:/var/www/acme",
		"/var/run/docker.sock:/var/run/docker.sock",
	},
}

// All box managed containers will start with this prefix
const boxContainerPrefix = "box__"

var devServices = []manifest.Service{
	routerService,
}
var prodServices = []manifest.Service{
	routerService,
	registryService,
	cronService,
}

const prodDataDir = "/mnt/data"

// New returns a new instance of the runtime structure for the supplied project
func New(mfst *manifest.Manifest, cfg *config.Config, isProduction bool) (*Runtime, error) {
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
	runFilename := filepath.Join(configDir, ".run.yml")
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

	ctx := context.Background()
	if runInfo.Project != mfst.Project {
		oldContainerIDs := []string{}

		// The previous running project was not this project, so verify that no containers are running
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
		Manifest:   mfst,
		Client:     cli,
		Context:    ctx,
		Production: isProduction,
		Config:     cfg,
	}, nil
}

func (rt *Runtime) Shutdown() error {
	defer rt.Client.Close()
	containers, err := rt.Client.ContainerList(rt.Context, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	containerIDs := []string{}
	for _, container := range containers {
		for _, name := range container.Names {
			if strings.HasPrefix(name, boxContainerPrefix) {
				containerIDs = append(containerIDs, container.ID)
				break
			}
		}
	}

	if len(containerIDs) > 0 {
		group := new(errgroup.Group)
		for _, containerID := range containerIDs {
			fmt.Printf("Stopping container %v...", containerID)
			group.Go(func() error {
				if err = rt.Client.ContainerStop(rt.Context, containerID, nil); err != nil {
					fmt.Printf("Error stopping container %v\n", containerID)
					return err
				}
				fmt.Printf("Stopped container %v\n", containerID)
				return nil
			})

		}

		if err := group.Wait(); err == nil {
			fmt.Println("Completed")
		} else {
			fmt.Println("An error occurred while closing one or more containers")
		}

		group = new(errgroup.Group)
		for _, containerID := range containerIDs {
			fmt.Printf("Removing container %v...", containerID)
			group.Go(func() error {
				if err = rt.Client.ContainerRemove(
					rt.Context,
					containerID,
					types.ContainerRemoveOptions{
						RemoveVolumes: false,
						RemoveLinks:   false,
						Force:         false,
					},
				); err != nil {
					fmt.Printf("Error removing container %v\n", containerID)
					return err
				}
				fmt.Printf("Removed container %v\n", containerID)
				return nil
			})

		}

		if err := group.Wait(); err == nil {
			fmt.Println("Completed")
		} else {
			fmt.Println("An error occurred while closing one or more containers")
		}
	} else {
		fmt.Printf("No box containers running, nothing to do.")
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}
	runFilename := filepath.Join(configDir, ".run.yml")
	err = os.Remove(runFilename)

	return err
}

// Start will create and start the required containers
func (rt *Runtime) Start() error {
	var coreServices []manifest.Service
	if rt.Production == true {
		coreServices = prodServices
	} else {
		coreServices = devServices
	}

	allServices := []manifest.Service{}
	for _, service := range coreServices {
		allServices = append(allServices, service)
	}
	for _, service := range rt.Manifest.Services {
		allServices = append(allServices, service)
	}

	containerBodyByServiceName := map[string]container.ContainerCreateCreatedBody{}
	var createErr error

	// Create all the services
	for _, service := range allServices {
		hostname := service.GetHostname()
		containerName := fmt.Sprint("%v%v", boxContainerPrefix, service.Name)

		// If a routing exists, new containers automatically get a "_1" suffix,
		// same goes for hostname
		if service.Routing.Path.Pattern != "" {
			hostname = fmt.Sprintf("%v_1", hostname)
			containerName = fmt.Sprintf("%v_1", containerName)
		}

		contConfig := container.Config{
			Hostname:     hostname,
			Env:          service.GetEnv(),
			Image:        service.GetImage(rt.Config.ProjectNameHash()),
			ExposedPorts: service.GetContainerPortSet(),
		}

		var dataDir string
		if rt.Production == true {
			dataDir = prodDataDir
		} else {
			dataDir, err := rt.Config.DataDir()
			if err != nil {
				return err
			}
			if _, err := os.Stat(dataDir); err != nil {
				return err
			}
		}
		mounts := service.GetHostMounts(dataDir)
		for _, mount := range mounts {
			if _, err := os.Stat(mount.Source); err != nil {
				fmt.Printf("Preparing host bind mount point: %v\n", mount.Source)
				// More than likely the mount point doesn't exist on the host, so make it
				// It will be owned be the user running this command
				err = os.MkdirAll(mount.Source, os.FileMode(755))
				if err != nil {
					return err
				}
			}
		}

		hostConfig := container.HostConfig{
			PortBindings: service.GetHostPortMap(),
			Mounts:       mounts,
		}

		containerBody, createErr := rt.Client.ContainerCreate(
			rt.Context,
			&contConfig,
			&hostConfig,
			nil,
			nil,
			containerName,
		)

		if createErr != nil {
			break
		}

		containerBodyByServiceName[service.Name] = containerBody
	}

	if createErr != nil {
		fmt.Println("Error creating containers")
		for _, containerBody := range containerBodyByServiceName {
			fmt.Println("Removing container ", containerBody.ID)
			rt.Client.ContainerRemove(rt.Context, containerBody.ID, types.ContainerRemoveOptions{})
		}
		return fmt.Errorf("An error occurred while creating containers")
	}

	serviceTranches, err := makeDependencyTranches(coreServices)
	if err != nil {
		return err
	}

	// Add the core services as the first tranche
	tranches := []tranche{}
	tranches = append(tranches, tranche(coreServices))
	for _, tranche := range serviceTranches {
		tranches = append(tranches, tranche)
	}

	for _, tranche := range tranches {

		// All containers in a tranche get started together, each in a separate goroutine.
		group := new(errgroup.Group)
		for _, service := range tranche {

			group.Go(func() error {
				fmt.Println("Starting container for ", service.Name)
				containerBody := containerBodyByServiceName[service.Name]
				err := rt.Client.ContainerStart(rt.Context, containerBody.ID, types.ContainerStartOptions{})
				return err
			})
		}

		if err := group.Wait(); err == nil {
			fmt.Println("Completed")
		} else {
			fmt.Println("An error occurred while starting containers, rolling back...")
			newErr := rt.Shutdown()
			if newErr != nil {
				fmt.Println("An error occurred while shutting down containers")
				fmt.Println(newErr)
			}
			return err
		}
	}

	return nil
}

func (rt *Runtime) Build() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("runtime.Build: %w", err)
	}

	for serviceName, service := range rt.Manifest.Services {
		if service.Build.Context != "" {
			fmt.Println("Building image for ", serviceName)
			contextPath, err := filepath.Abs(filepath.Join(dir, service.Build.Context))
			if err != nil {
				return fmt.Errorf("runtime.Build: %w", err)
			}
			dockerfilePath := filepath.Join(contextPath, service.Build.Dockerfile)
			dockerfileAbsPath, err := filepath.Abs(dockerfilePath)
			if err != nil {
				return fmt.Errorf("runtime.Build: %w", err)
			}

			image := service.GetImage(rt.Config.ProjectNameHash())
			builder, err := cmd.New(
				"docker",
				"build",
				"--file", dockerfileAbsPath,
				"--tag", image,
			)
			if err != nil {
				return fmt.Errorf("runtime.Build: %w", err)
			}

			builder.CWD = contextPath
			err = builder.Run()
			if err != nil {
				return fmt.Errorf("Error building %v: %w", serviceName, err)
			}
		}
	}

	return nil
}
