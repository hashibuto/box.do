package main

import (
	"box/api/digitalocean"
	"box/api/digitalocean/action"
	"box/api/digitalocean/droplet"
	dropletenum "box/api/digitalocean/enum/droplet"
	"box/config"
	"box/sshconn"
	"fmt"
	"os"
	"time"
)

const imageHostName = "box-image-maker"
const maxConnectAttempts = 10
const sshRetrySeconds = 10
const configRepo = "box.do-config"
const adminUser = "owner"

type MkImageCmd struct {
	Name      string `arg help="Project name"`
	Overwrite bool   `default=false help="Overwrite existing image"`
}

// Run runs the make image command, which creates a droplet image with the Box base confugration
// for use in subsequent droplet deployments.
func (cmd *MkImageCmd) Run() error {
	var dropletObj *droplet.Droplet
	cfg, err := config.Load(cmd.Name)
	if err != nil {
		return err
	}

	if cmd.Overwrite == false && cfg.ImageID != 0 {
		fmt.Println("It appears you've already created a DigitalOcean image.")
		fmt.Println("Please use the --overwrite option if you'd like to overwrite it with a new one.")
		os.Exit(1)
	}

	// Obtain the SSH signer first in order to avoid creating unnecessary resources, should this process fail
	signer, err := sshconn.GetSigner(cfg.PrivateKeyFilename)
	if err != nil {
		return err
	}

	fmt.Print("Creating new droplet for base image...")
	doSvc := digitalocean.NewService(cfg.DigitalOceanAPIKey)
	dropletObj, err = droplet.CreateFromPublicImage(
		doSvc,
		imageHostName,
		dropletenum.S1VCPU1GB, // Minimum size for image build
		cfg.Region,
		droplet.DefaultPublicImage,
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

	connected := false
	var conn *sshconn.SSHConn
	for i := 0; i < maxConnectAttempts; i++ {
		fmt.Println("Trying to contact via SSH...")
		conn, err = sshconn.NewSSHConn(signer, "root", ipAddress)
		if err == nil {
			connected = true
			break
		}
		fmt.Println(err)
		fmt.Printf("Failed, trying again in %vs...\n", sshRetrySeconds)
		time.Sleep(time.Second * sshRetrySeconds)
	}

	if connected == false {
		fmt.Printf("Unable to connect via SSH, aborting")
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("SSH connection established, continuing")

	err = conn.Run([]string{
		fmt.Sprintf("wget -O /root/config_image.sh https://raw.githubusercontent.com/hashibuto/%v/master/scripts/config_image.sh", configRepo),
		"chmod +x /root/config_image.sh",
		fmt.Sprintf("/root/config_image.sh %v %v", adminUser, configRepo),
	})
	if err != nil {
		return err
	}

	fmt.Printf("Waiting for droplet to power down...")
	for dropletObj.Status != "off" {
		time.Sleep(time.Second * 5)
		fmt.Print("Checking droplet status...")
		dropletObj, err = droplet.Get(doSvc, dropletObj.ID)
		if err != nil {
			return err
		}
		fmt.Println(dropletObj.Status)
	}

	fmt.Println("Droplet powered down, creating snapshot")

	actionObj, err := droplet.CreateSnapshot(doSvc, dropletObj.ID, "box-base")
	if err != nil {
		return err
	}

	for actionObj.Status != "completed" {
		time.Sleep(time.Second * 5)
		fmt.Print("Checking action status...")
		actionObj, err = action.Get(doSvc, actionObj.ID)
		if err != nil {
			return err
		}
		fmt.Println(actionObj.Status)
	}

	fmt.Println("Snapshot complete")

	dropletObj, err = droplet.Get(doSvc, dropletObj.ID)
	if err != nil {
		return err
	}

	cfg.ImageID = dropletObj.SnapshotIds[0]
	err = cfg.Save()
	if err != nil {
		return err
	}

	fmt.Print("Deleting droplet...")
	err = droplet.Delete(doSvc, dropletObj.ID)
	if err != nil {
		return err
	}
	fmt.Println("Done")
	return err
}
