package droplet

import (
	"box/api/digitalocean"
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
	ID       int    `json:"id"`
	Status   string `json:"status"`
	Networks struct {
		V4 []Address `json:"v4"`
	} `json:"networks"`
}

type createFromPublicImageRequest struct {
	Name    string   `json:"name"`
	Size    string   `json:"size"`
	Region  string   `json:"region"`
	Image   string   `json:"image"`
	SSHKeys []string `json:"ssh_keys"`
}

type createFromPrivateImageRequest struct {
	Name    string   `json:"name"`
	Size    string   `json:"size"`
	Region  string   `json:"region"`
	Image   int      `json:"image"`
	SSHKeys []string `json:"ssh_keys"`
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
func CreateFromPublicImage(svc *digitalocean.Service, name, size, region, imageSlug string, SSHKeys []int) (*Droplet, error) {
	cr := createFromPublicImageRequest{
		Name:   name,
		Size:   size,
		Region: region,
		Image:  imageSlug,
	}

	reqBody, err := json.Marshal(&cr)
	if err != nil {
		return nil, err
	}

	return create(svc, reqBody)
}

// CreateFromPrivateImage creates a droplet from a public image ID
func CreateFromPrivateImage(svc *digitalocean.Service, name, size, region string, imageID int, SSHKeys []int) (*Droplet, error) {
	cr := createFromPrivateImageRequest{
		Name:   name,
		Size:   size,
		Region: region,
		Image:  imageID,
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
