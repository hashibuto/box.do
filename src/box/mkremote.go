package main

import (
	"box/api/digitalocean"
	"box/api/digitalocean/domain"
	"box/config"
	"fmt"
	"os"
)

type MakeRemoteCmd struct {
	Name string `arg help="Project name"`
}

var doNameServers []string = []string{
	"ns1.digitalocean.com",
	"ns2.digitalocean.com",
	"ns3.digitalocean.com",
}

func (cmd *MakeRemoteCmd) Run() error {
	cfg, err := config.Load(cmd.Name)
	if err != nil {
		return err
	}

	if cfg.DropletID != 0 {
		fmt.Println("There is already a droplet ID associated with this project. If you want to make a new remote host, please delete this droplet first.")
		os.Exit(1)
	}
	doSvc := digitalocean.NewService(cfg.DigitalOceanAPIKey)

	// Check that the bare domain is present in DigitalOcean's network section
	domainObj, err := domain.Get(doSvc, cfg.BareDomainName)
	if err != nil {
		if e, ok := err.(*digitalocean.RespError); ok {
			if e.StatusCode != 404 {
				return err
			}
		} else {
			return err
		}
	}

	if domainObj == nil {
		fmt.Println("Please ensure that your domain registrar points your domain to digitalocean's nameservers")
		for _, ns := range doNameServers {
			fmt.Println(" - ", ns)
		}

		fmt.Printf("Creating domain entry for %v...", cfg.BareDomainName)
		_, err := domain.Create(doSvc, cfg.BareDomainName)
		if err != nil {
			return nil
		}
		fmt.Println("Done")
	} else {
		fmt.Println("Domain entry exists for", cfg.BareDomainName)
	}

	return nil
}
