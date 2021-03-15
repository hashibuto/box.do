package blockstorage

import (
	"box/api/digitalocean"
	"box/api/digitalocean/action"
	"encoding/json"
	"fmt"
)

type Volume struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type volumeResp struct {
	Volume Volume `json:"volume"`
}

const basePath = "/volumes"

// Get retrieves a volume by ID
func Get(svc *digitalocean.Service, ID string) (*Volume, error) {
	respBody, err := svc.Get(fmt.Sprintf("%v/%v", basePath, ID))
	if err != nil {
		return nil, err
	}

	getResp := volumeResp{}
	err = json.Unmarshal(respBody, &getResp)
	if err != nil {
		return nil, err
	}

	return &getResp.Volume, nil
}

// Create provisions a block storage volume in the supplied region
func Create(svc *digitalocean.Service, name, region string, sizeInGb int) (*Volume, error) {
	type createReq struct {
		SizeGigabytes int    `json:"size_gigabytes"`
		Name          string `json:"name"`
		Region        string `json:"region"`
	}

	create := createReq{
		SizeGigabytes: sizeInGb,
		Name:          name,
		Region:        region,
	}

	reqBody, err := json.Marshal(&create)
	if err != nil {
		return nil, err
	}

	respBody, err := svc.Post(basePath, reqBody)
	if err != nil {
		return nil, err
	}

	createResp := volumeResp{}
	err = json.Unmarshal(respBody, &createResp)
	if err != nil {
		return nil, err
	}

	return &createResp.Volume, nil
}

// Attach attaches an existing block storage volume to an existing droplet
func Attach(svc *digitalocean.Service, volumeID string, dropletID int) (*action.Action, error) {
	type attachReq struct {
		Type      string `json:"type"`
		DropletID int    `json:"droplet_id"`
	}
	attach := attachReq{
		Type:      "attach",
		DropletID: dropletID,
	}
	reqBody, err := json.Marshal(&attach)
	if err != nil {
		return nil, err
	}
	respBody, err := svc.Post(fmt.Sprintf("%v/%v/actions", basePath, volumeID), reqBody)
	if err != nil {
		return nil, err
	}

	actionResp := action.Response{}
	err = json.Unmarshal(respBody, &actionResp)
	if err != nil {
		return nil, err
	}

	return &actionResp.Action, nil
}
