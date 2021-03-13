package domain

import (
	"box/api/digitalocean"
	"encoding/json"
	"fmt"
)

type Domain struct {
	Name string `json:"name"`
}

type domainResp struct {
	Domain Domain `json:"domain"`
}

const basePath = "/domains"

// Get retrieves a domain by name if one is managed within the account, or returns an error
func Get(svc *digitalocean.Service, domainName string) (*Domain, error) {
	respBody, err := svc.Get(fmt.Sprintf("%v/%v", basePath, domainName))
	if err != nil {
		return nil, err
	}

	getResp := domainResp{}
	err = json.Unmarshal(respBody, &getResp)
	if err != nil {
		return nil, err
	}

	return &getResp.Domain, nil
}

// Create adds domainName to the list of managed domains within the digitalocean account
func Create(svc *digitalocean.Service, domainName string) (*Domain, error) {
	type createReq struct {
		Name string `json:"name"`
	}
	req := createReq{
		Name: domainName,
	}
	reqBody, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	respBody, err := svc.Post(basePath, reqBody)
	createResp := domainResp{}
	err = json.Unmarshal(respBody, &createResp)
	if err != nil {
		return nil, err
	}

	return &createResp.Domain, nil
}
