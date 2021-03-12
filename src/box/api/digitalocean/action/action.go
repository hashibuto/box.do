package action

import (
	"box/api/digitalocean"
	"encoding/json"
	"fmt"
)

const basePath = "/actions"

// Action data structure
type Action struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

// Response represents a standard action response structure
type Response struct {
	Action Action `json:"action"`
}

// Get retrieves an action by ID
func Get(svc *digitalocean.Service, actionID int) (*Action, error) {
	respBody, err := svc.Get(fmt.Sprintf("%v/%v", basePath, actionID))
	if err != nil {
		return nil, err
	}

	actionResp := &Response{}
	err = json.Unmarshal(respBody, actionResp)
	if err != nil {
		return nil, err
	}

	return &actionResp.Action, nil
}
