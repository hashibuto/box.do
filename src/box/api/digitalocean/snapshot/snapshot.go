package snapshot

import (
	"box/api/digitalocean"
	"encoding/json"
	"fmt"
)

const basePath = "/snapshots"

type Snapshot struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetAllDropletSnapshots returns up to 200 snapshots created from droplets (pagination
// is not supported, a maximum of 200 rows can be retrieved)
func GetAllDropletSnapshots(svc *digitalocean.Service) ([]Snapshot, error) {
	respBody, err := svc.Get(fmt.Sprintf("%v?resource_type=droplet&per_page=200", basePath))
	if err != nil {
		return nil, fmt.Errorf("GetAllDropletSnapshots: %w", err)
	}

	getResp := struct {
		Snapshots []Snapshot `json:"snapshots"`
	}{}

	err = json.Unmarshal(respBody, &getResp)
	if err != nil {
		return nil, fmt.Errorf("GetAllDropletSnapshots: %w", err)
	}

	return getResp.Snapshots, nil
}

// Delete deletes a snapshot by ID
func Delete(svc *digitalocean.Service, snapshotID string) error {
	return svc.Delete(fmt.Sprintf("%v/%v", basePath, snapshotID))
}
