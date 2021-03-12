package droplet

import (
	"box/api/digitalocean"
	"box/api/digitalocean/action"
	"encoding/json"
	"fmt"
)

// DefaultPublicImage is the public image slug upon which to build base images
const DefaultPublicImage = "ubuntu-20-04-x64"

// Address represents the IP address and type of address for a given network where types are "public" or "private"
type Address struct {
	Type      string `json:"type"`
	IPAddress string `json:"ip_address"`
}

// Droplet represents a droplet structure
type Droplet struct {
	ID          int    `json:"id"`
	Status      string `json:"status"`
	SnapshotIds []int  `json:"snapshot_ids"`
	Networks    struct {
		V4 []Address `json:"v4"`
	} `json:"networks"`
}

type createFromPublicImageRequest struct {
	Name    string `json:"name"`
	Size    string `json:"size"`
	Region  string `json:"region"`
	Image   string `json:"image"`
	SSHKeys []int  `json:"ssh_keys"`
}

type createFromPrivateImageRequest struct {
	Name    string `json:"name"`
	Size    string `json:"size"`
	Region  string `json:"region"`
	Image   int    `json:"image"`
	SSHKeys []int  `json:"ssh_keys"`
}

type createResponse struct {
	Droplet Droplet `json:"droplet"`
}

const basePath = "/droplets"

func create(svc *digitalocean.Service, reqBody []byte) (*Droplet, error) {
	respBody, err := svc.Post(basePath, reqBody)
	if err != nil {
		return nil, err
	}

	respObj := createResponse{}
	err = json.Unmarshal(respBody, &respObj)
	if err != nil {
		return nil, err
	}

	return &respObj.Droplet, nil
}

// CreateFromPublicImage creates a droplet from a public image slug
func CreateFromPublicImage(svc *digitalocean.Service, name, size, region, imageSlug string, sshKeys []int) (*Droplet, error) {
	cr := createFromPublicImageRequest{
		Name:    name,
		Size:    size,
		Region:  region,
		Image:   imageSlug,
		SSHKeys: sshKeys,
	}

	reqBody, err := json.Marshal(&cr)
	if err != nil {
		return nil, err
	}

	return create(svc, reqBody)
}

// CreateFromPrivateImage creates a droplet from a public image ID
func CreateFromPrivateImage(svc *digitalocean.Service, name, size, region string, imageID int, sshKeys []int) (*Droplet, error) {
	cr := createFromPrivateImageRequest{
		Name:    name,
		Size:    size,
		Region:  region,
		Image:   imageID,
		SSHKeys: sshKeys,
	}

	reqBody, err := json.Marshal(&cr)
	if err != nil {
		return nil, err
	}

	return create(svc, reqBody)
}

// GetByID returns a droplet object for the provided droplet ID
func GetByID(svc *digitalocean.Service, dropletID int) (*Droplet, error) {
	respBody, err := svc.Get(fmt.Sprintf("%v/%v", basePath, dropletID))
	if err != nil {
		return nil, err
	}

	respObj := createResponse{}
	err = json.Unmarshal(respBody, &respObj)
	if err != nil {
		return nil, err
	}

	return &respObj.Droplet, nil
}

// CreateSnapshot creates a named snapshot of the given droplet
func CreateSnapshot(svc *digitalocean.Service, dropletID int, name string) (*action.Action, error) {
	type createSnapshotRequest struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}

	req := &createSnapshotRequest{
		Type: "snapshot",
		Name: name,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	respBody, err := svc.Post(fmt.Sprintf("%v/%v/actions", basePath, dropletID), reqBody)
	if err != nil {
		return nil, err
	}

	actionResp := &action.Response{}
	err = json.Unmarshal(respBody, actionResp)
	if err != nil {
		return nil, err
	}

	return &actionResp.Action, nil
}

// Delete deletes a droplet by ID
func Delete(svc *digitalocean.Service, dropletID int) error {
	return svc.Delete(fmt.Sprintf("%v/%v", basePath, dropletID))
}
