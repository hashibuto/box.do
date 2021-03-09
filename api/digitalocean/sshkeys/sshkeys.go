package sshkeys

import (
	"encoding/json"
	"prdo/api/digitalocean"
)

type sshKey struct {
	ID int `json:"id"`
}

type createBody struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type createResponse struct {
	SSHKey sshKey `json:"ssh_key"`
}

const basePath = "/account/keys"

// Create creates an SSH key entry for publicKey and returns the ID if successful
func Create(svc *digitalocean.Service, name string, publicKey string) (int, error) {
	body := &createBody{
		Name:      name,
		PublicKey: publicKey,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	respBody, err := svc.Post(basePath, jsonBody)
	if err != nil {
		return 0, err
	}

	createdKey := &createResponse{}
	err = json.Unmarshal(respBody, createdKey)
	if err != nil {
		return 0, err
	}

	return createdKey.SSHKey.ID, nil
}
