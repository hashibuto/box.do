package manifest

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
)

const LocalImagePrefix = "@/"

type BuildInfo struct {
	Context    string `yaml:"context"`
	Dockerfile string `yaml:"dockerfile"`
}

type Service struct {
	Name        string
	Hostname    string            `yaml:"hostname"`
	Routing     Routing           `yaml:"routing"`
	Environment map[string]string `yaml:"environment"`
	Volumes     []string          `yaml:"volumes"`
	Image       string            `yaml:"image"`
	DependsOn   []string          `yaml:"depends_on"`
	Ports       []string          `yaml:"ports"`
	Build       BuildInfo         `yaml:"build"`
}

func (svc *Service) GetHostname() string {
	if svc.Hostname != "" {
		return svc.Hostname
	}

	return svc.Name
}

func (svc *Service) GetEnv() []string {
	envVars := []string{}
	for key, value := range svc.Environment {
		envVars = append(envVars, fmt.Sprintf("%v=%v", key, value))
	}

	return envVars
}

// GetImage returns the image, replacing any local reference with a unique project identifier
func (svc *Service) GetImage(uniqueSuffix string) string {
	var imgStr string
	if strings.HasPrefix(svc.Image, LocalImagePrefix) {
		imgStr = fmt.Sprintf("%v_%v", svc.Image[len(LocalImagePrefix):], uniqueSuffix)
	} else {
		imgStr = svc.Image
	}

	if !strings.Contains(imgStr, ":") {
		imgStr = fmt.Sprintf("%v:latest", imgStr)
	}

	return imgStr
}

func (svc *Service) GetContainerPortSet() nat.PortSet {
	portSet := nat.PortSet{}

	for _, portStr := range svc.Ports {
		portParts := strings.Split(portStr, ":")
		containerPort := portParts[1]
		portSet[nat.Port(fmt.Sprintf("%v/tcp", containerPort))] = struct{}{}
	}

	return portSet
}

func (svc *Service) GetHostPortMap() nat.PortMap {
	portMap := nat.PortMap{}

	for _, portStr := range svc.Ports {
		portParts := strings.Split(portStr, ":")
		hostPort := portParts[0]
		containerPort := portParts[1]
		portMap[nat.Port(fmt.Sprintf("%v/tcp", containerPort))] = []nat.PortBinding{
			{
				// All bindings are to localhost.  We don't allow binding to
				// 0.0.0.0 since all additional outside connections should be tunneled through
				// port 22
				HostIP:   "127.0.0.1",
				HostPort: hostPort,
			},
		}
	}

	return portMap
}

func (svc *Service) GetHostMounts(dataDir string) []mount.Mount {
	mounts := []mount.Mount{}
	for _, volume := range svc.Volumes {
		volumeParts := strings.Split(volume, ":")
		hostPath := volumeParts[0]
		if strings.HasPrefix(hostPath, LocalImagePrefix) {
			hostPath = filepath.Join(dataDir, hostPath[len(LocalImagePrefix):])
		}
		containerPath := volumeParts[1]
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: hostPath,
			Target: containerPath,
		})
	}

	return mounts
}
