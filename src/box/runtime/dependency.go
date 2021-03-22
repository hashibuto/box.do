package runtime

import (
	"box/manifest"
	"fmt"
)

type tranche []manifest.Service

func makeDependencyTranches(services []manifest.Service) ([]tranche, error) {
	// Copy the map to preserve the original
	remainingServices := map[string]manifest.Service{}
	for _, service := range services {
		remainingServices[service.Name] = service
	}

	tranches := []tranche{}
	for len(remainingServices) > 0 {
		current := []manifest.Service{}
		for serviceName, service := range remainingServices {
			remainingDeps := false
			for _, dep := range service.DependsOn {
				if _, ok := remainingServices[dep]; ok {
					remainingDeps = true
					break
				}
			}
			if remainingDeps == false {
				current = append(current, service)
				delete(remainingServices, serviceName)
			}
		}

		if len(current) == 0 {
			return nil, fmt.Errorf("Impossible depends_on tree in services")
		}

		tranches = append(tranches, current)
	}

	return tranches, nil
}
