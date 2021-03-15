package domain

import (
	"box/api/digitalocean"
	"encoding/json"
	"fmt"
)

type Domain struct {
	Name string `json:"name"`
}

type DomainRecord struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
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

// ListRecords will list all records of a given type for the provided domain name.
// For the sake of simplicity, a maximum of 200 records can be returned.  Pagination
// is not supported.
func ListRecords(svc *digitalocean.Service, domainName string, filter string) ([]DomainRecord, error) {
	url := fmt.Sprintf("%v/%v/records?per_page=200", basePath, domainName)
	if filter != "" {
		url = fmt.Sprintf("%v&type=%v", url, filter)
	}

	respBody, err := svc.Get(url)
	if err != nil {
		return nil, err
	}

	type listResp struct {
		DomainRecords []DomainRecord `json:"domain_records"`
	}

	resp := listResp{}
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return nil, err
	}

	return resp.DomainRecords, nil
}

// DeleteRecord deletes a domain record belonging to the named domain
func DeleteRecord(svc *digitalocean.Service, domainName string, recordID int) error {
	return svc.Delete(fmt.Sprintf("%v/%v/records/%v", basePath, domainName, recordID))
}

// Creates an A record for the provided domain, which points to the IP address
func CreateARecord(svc *digitalocean.Service, domainName, name, ipAddress string, ttlSec int) (*DomainRecord, error) {
	type createARecord struct {
		Type string `json:"type"`
		Name string `json:"name"`
		Data string `json:"data"`
		TTL  int    `json:"ttl"`
	}

	create := createARecord{
		Type: "A",
		Name: name,
		Data: ipAddress,
		TTL:  ttlSec,
	}

	reqBody, err := json.Marshal(&create)
	if err != nil {
		return nil, err
	}

	respBody, err := svc.Post(fmt.Sprintf("%v/%v/records", basePath, domainName), reqBody)
	if err != nil {
		return nil, err
	}

	type domainRecResp struct {
		DomainRecord DomainRecord `json:"domain_record"`
	}

	createResp := domainRecResp{}
	err = json.Unmarshal(respBody, &createResp)
	if err != nil {
		return nil, err
	}

	return &createResp.DomainRecord, nil
}
