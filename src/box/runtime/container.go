package runtime

import (
	"box/manifest"
	"fmt"
	"os"

	"github.com/docker/docker/api/types/container"
)

// CreateContainer creates a container using the Runtime object, from a provided Service manifest
func (rt *Runtime) CreateContainer(service *manifest.Service) (*container.ContainerCreateCreatedBody, error) {
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
			return nil, err
		}
		if _, err := os.Stat(dataDir); err != nil {
			return nil, fmt.Errorf("Unable to stat data directory %v: %w", dataDir, err)
		}
	}
	mounts := service.GetHostMounts(dataDir)
	for _, mount := range mounts {
		if _, err := os.Stat(mount.Source); err != nil {
			// More than likely the mount point doesn't exist on the host, so make it
			// It will be owned be the user running this command
			err = os.MkdirAll(mount.Source, os.FileMode(755))
			if err != nil {
				return nil, fmt.Errorf("Unable to make host mount directory %v: %w", mount.Source, err)
			}
		}
	}

	hostConfig := container.HostConfig{
		PortBindings: service.GetHostPortMap(),
		Mounts:       mounts,
	}

	containerBody, err := rt.Client.ContainerCreate(
		rt.Context,
		&contConfig,
		&hostConfig,
		nil,
		nil,
		containerName,
	)
	if err != nil {
		return nil, fmt.Errorf("Create container failed: %w", err)
	}

	return &containerBody, nil
}
