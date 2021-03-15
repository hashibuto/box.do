package main

import (
	"box/api/digitalocean"
	"box/api/digitalocean/action"
	"box/api/digitalocean/blockstorage"
	"box/api/digitalocean/domain"
	"box/api/digitalocean/droplet"
	"box/api/digitalocean/firewall"
	"box/config"
	"fmt"
	"os"
	"strings"
	"time"
)

type MakeRemoteCmd struct {
	Name string `arg help="Project name"`
}

var doNameServers []string = []string{
	"ns1.digitalocean.com",
	"ns2.digitalocean.com",
	"ns3.digitalocean.com",
}

var allAddresses []string = []string{
	"0.0.0.0/0",
	"::/0",
}

func (cmd *MakeRemoteCmd) Run() error {
	cfg, err := config.Load(cmd.Name)
	if err != nil {
		return err
	}

	if cfg.ImageID == 0 {
		fmt.Println("Please build the box deployment image first using: box mkimage")
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

	createFirewall := false
	if cfg.FirewallID != "" {
		// Verify the presence of the existing firewall, if it doesn't exist, it will be replaced
		fmt.Print("Verifying existing firewall...")
		_, err := firewall.Get(doSvc, cfg.FirewallID)
		if err != nil {
			if e, ok := err.(*digitalocean.RespError); ok {
				if e.StatusCode == 404 {
					fmt.Println("Not found")
					createFirewall = true
				} else {
					return e
				}
			} else {
				return err
			}
		} else {
			fmt.Println("Found")
		}
	} else {
		fmt.Println("No firewall configured")
		createFirewall = true
	}

	if createFirewall == true {
		fmt.Print("Creating firewall...")
		name := fmt.Sprintf("box-%v", strings.ToLower(cfg.ProjectName))

		inboundRules := []firewall.InboundRule{
			{
				Protocol: "tcp",
				Ports:    "22",
				Sources: firewall.Addresses{
					Addresses: allAddresses,
				},
			},
			{
				Protocol: "tcp",
				Ports:    "80",
				Sources: firewall.Addresses{
					Addresses: allAddresses,
				},
			},
			{
				Protocol: "tcp",
				Ports:    "443",
				Sources: firewall.Addresses{
					Addresses: allAddresses,
				},
			},
			{
				Protocol: "icmp",
				Sources: firewall.Addresses{
					Addresses: allAddresses,
				},
			},
		}

		outboundRules := []firewall.OutboundRule{
			{
				Protocol: "icmp",
				Destinations: firewall.Addresses{
					Addresses: allAddresses,
				},
			},
			{
				Protocol: "tcp",
				Ports:    "all",
				Destinations: firewall.Addresses{
					Addresses: allAddresses,
				},
			},
			{
				Protocol: "udp",
				Ports:    "all",
				Destinations: firewall.Addresses{
					Addresses: allAddresses,
				},
			},
		}

		firewallObj, err := firewall.Create(doSvc, name, inboundRules, outboundRules)
		if err != nil {
			return err
		}
		fmt.Println("Done")

		cfg.FirewallID = firewallObj.ID

		// Save in order to prevent redoing this step if a failure occurs
		err = cfg.Save()
		if err != nil {
			return err
		}
	}

	// Check if block storage has been provisioned, if not provision it
	createVolume := false
	if cfg.BlockStorageID != "" {
		fmt.Print("Verifying existing block storage volume...")
		_, err := blockstorage.Get(doSvc, cfg.BlockStorageID)
		if err != nil {
			if e, ok := err.(*digitalocean.RespError); ok {
				if e.StatusCode == 404 {
					// Volume doesn't exist
					createVolume = true
					fmt.Println("Not found")
				}
			} else {
				return err
			}
		} else {
			fmt.Println("Found")
		}
	} else {
		fmt.Println("No block storage volume configured")
		createVolume = true
	}

	if createVolume == true {
		fmt.Print("Creating block storage volume...")
		name := fmt.Sprintf("box-%v", strings.ToLower(cfg.ProjectName))
		volumeObj, err := blockstorage.Create(doSvc, name, cfg.Region, cfg.VolumeSize)
		if err != nil {
			return err
		}

		cfg.BlockStorageID = volumeObj.ID
		// Save in order to prevent redoing this step if a failure occurs
		err = cfg.Save()
		if err != nil {
			return err
		}
		fmt.Println("Done")
	}

	var dropletObj *droplet.Droplet
	// Check if droplet has been provisioned, if not provision it
	createDroplet := false
	if cfg.DropletID != 0 {
		fmt.Print("Verifying existing droplet...")
		dropletObj, err = droplet.Get(doSvc, cfg.DropletID)
		if err != nil {
			if e, ok := err.(*digitalocean.RespError); ok {
				if e.StatusCode == 404 {
					// Droplet doesn't exist
					createDroplet = true
					fmt.Println("Not found")
				}
			} else {
				return err
			}
		} else {
			fmt.Println("Found")
		}
	} else {
		fmt.Println("No droplet configured")
		createDroplet = true
	}

	if createDroplet == true {
		fmt.Print("Creating droplet...")
		name := fmt.Sprintf("box-%v", strings.ToLower(cfg.ProjectName))
		dropletObj, err = droplet.CreateFromPrivateImage(
			doSvc,
			name,
			cfg.DropletSlug,
			cfg.Region,
			cfg.ImageID,
			[]int{cfg.PublicKeyID},
		)
		if err != nil {
			return err
		}
		fmt.Println("Done")

		for dropletObj.Status == "new" {
			time.Sleep(time.Second * 5)
			fmt.Print("Checking droplet status...")
			dropletObj, err = droplet.Get(doSvc, dropletObj.ID)
			if err != nil {
				return err
			}
			fmt.Println(dropletObj.Status)
		}

		if dropletObj.Status != "active" {
			fmt.Printf("Droplet creation failed, please be sure to remove the droplet named %v from your DigitalOcean control panel\n", imageHostName)
			os.Exit(1)
		}

		fmt.Println("Droplet successfully created")

		var ipAddress string
		for _, address := range dropletObj.Networks.V4 {
			if address.Type == "public" {
				ipAddress = address.IPAddress
			}
		}

		if ipAddress == "" {
			fmt.Println("Unable to obtain the public IP4 address from the droplet object, aborting")
			os.Exit(1)
		}

		cfg.DropletID = dropletObj.ID
		cfg.DropletPublicIP = ipAddress

		// Save in order to prevent redoing this step if a failure occurs
		err = cfg.Save()
		if err != nil {
			return err
		}
	}

	found := false
	for _, volumeID := range dropletObj.VolumeIds {
		if volumeID == cfg.BlockStorageID {
			found = true
			break
		}
	}

	if found == true {
		fmt.Println("Block storage volume is already attached to droplet")
	} else {
		fmt.Print("Attaching block storage volume to droplet...")
		actionObj, err := blockstorage.Attach(doSvc, cfg.BlockStorageID, cfg.DropletID)
		if err != nil {
			return err
		}
		fmt.Println("Started")

		for actionObj.Status != "completed" {
			time.Sleep(time.Second * 5)
			fmt.Print("Checking action status...")
			actionObj, err = action.Get(doSvc, actionObj.ID)
			if err != nil {
				return err
			}
			fmt.Println(actionObj.Status)
		}

		fmt.Println("Block storage volume attached to droplet")
	}

	// Make sure an A domain record exists for the droplet
	fmt.Print("Verifying domain record...")
	domainRecs, err := domain.ListRecords(doSvc, cfg.BareDomainName, "A")
	if err != nil {
		return err
	}

	// Looking for an @ record, which indicates the bare domain
	var matchingDomainRec *domain.DomainRecord
	for _, domainRec := range domainRecs {
		if domainRec.Name == "@" {
			matchingDomainRec = &domainRec
		}
	}

	// If the @ record is pointing to the current droplet, nothing more to do
	// otherwise it needs to be deleted
	if matchingDomainRec != nil {
		if matchingDomainRec.Data != cfg.DropletPublicIP {
			// Needs to be deleted
			fmt.Println("Mismatch")
			fmt.Println("Deleting existing domain record")
			err = domain.DeleteRecord(doSvc, cfg.BareDomainName, matchingDomainRec.ID)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Match")
		}
	} else {
		fmt.Println("Not found")
		fmt.Printf("Adding domain record for %v...", cfg.DropletPublicIP)
		_, err := domain.CreateARecord(doSvc, cfg.BareDomainName, "@", cfg.DropletPublicIP, 1800)
		if err != nil {
			return err
		}
		fmt.Println("Done")
	}

	return nil
}
