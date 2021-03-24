package sshkeys

import (
	"box/api/digitalocean"
	"encoding/json"
	"fmt"
)

type SSHKey struct {
	ID        int    `json:"id"`
	PublicKey string `json:"public_key"`
}

type createBody struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type createResponse struct {
	SSHKey SSHKey `json:"ssh_key"`
}

const basePath = "/account/keys"

// Create creates an SSH key entry for publicKey and returns the ID if successful
func Create(svc *digitalocean.Service, name string, publicKey string) (*SSHKey, error) {
	body := &createBody{
		Name:      name,
		PublicKey: publicKey,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	respBody, err := svc.Post(basePath, jsonBody)
	if err != nil {
		return nil, err
	}

	createdKey := createResponse{}
	err = json.Unmarshal(respBody, &createdKey)
	if err != nil {
		return nil, err
	}

	return &createdKey.SSHKey, nil
}

// GetAll retrieves all SSH keys in the account.  Lists up to 200 results.
func GetAll(svc *digitalocean.Service) ([]SSHKey, error) {
	respBody, err := svc.Get(fmt.Sprintf("%v?per_page=200", basePath))
	if err != nil {
		return nil, fmt.Errorf("sshkeys.GetAll: %w", err)
	}

	keys := struct {
		SSHKeys []SSHKey `json:"ssh_keys"`
	}{}
	err = json.Unmarshal(respBody, &keys)
	if err != nil {
		return nil, fmt.Errorf("sshkeys.GetAll: %w", err)
	}

	return keys.SSHKeys, nil
}
