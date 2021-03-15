package firewall

import (
	"box/api/digitalocean"
	"encoding/json"
	"fmt"
)

type Firewall struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Name   string `json:"name"`
}

type Addresses struct {
	Addresses []string `json:"addresses"`
}

type InboundRule struct {
	Protocol string    `json:"protocol"`
	Ports    string    `json:"ports"`
	Sources  Addresses `json:"sources"`
}

type OutboundRule struct {
	Protocol     string    `json:"protocol"`
	Ports        string    `json:"ports"`
	Destinations Addresses `json:"destinations"`
}

type firewallResp struct {
	Firewall Firewall `json:"firewall"`
}

const basePath = "/firewalls"

// Get retrieves a firewall object by the provided ID
func Get(svc *digitalocean.Service, ID string) (*Firewall, error) {
	respBody, err := svc.Get(fmt.Sprintf("%v/%v", basePath, ID))
	if err != nil {
		return nil, err
	}

	getResp := firewallResp{}
	err = json.Unmarshal(respBody, &getResp)
	if err != nil {
		return nil, err
	}

	return &getResp.Firewall, nil
}

// Create will create a named firewall with the provided inbound and outbound ruleset
func Create(svc *digitalocean.Service, name string, inboundRules []InboundRule, outboundRules []OutboundRule) (*Firewall, error) {
	type createReq struct {
		Name          string         `json:"name"`
		InboundRules  []InboundRule  `json:"inbound_rules"`
		OutboundRules []OutboundRule `json:"outbound_rules"`
	}

	create := createReq{
		Name:          name,
		InboundRules:  inboundRules,
		OutboundRules: outboundRules,
	}

	reqBody, err := json.Marshal(&create)
	if err != nil {
		return nil, err
	}

	respBody, err := svc.Post(basePath, reqBody)
	if err != nil {
		return nil, err
	}

	createResp := firewallResp{}
	err = json.Unmarshal(respBody, &createResp)
	if err != nil {
		return nil, err
	}

	return &createResp.Firewall, nil
}
